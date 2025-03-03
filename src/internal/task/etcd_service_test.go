package task

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	taskapi "github.com/pachyderm/pachyderm/v2/src/task"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"golang.org/x/sync/errgroup"
)

var (
	errTaskFailure = errors.Errorf("task failure")
)

func serializeTestTask(testTask *TestTask) (*types.Any, error) {
	serializedTestTask, err := proto.Marshal(testTask)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + string(proto.MessageName(testTask)),
		Value:   serializedTestTask,
	}, nil
}

func deserializeTestTask(any *types.Any) (*TestTask, error) {
	testTask := &TestTask{}
	if err := types.UnmarshalAny(any, testTask); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return testTask, nil
}

func newTestEtcdService(t *testing.T) Service {
	env := testetcd.NewEnv(t)
	return NewEtcdService(env.EtcdClient, "")
}

func seedRand() string {
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
}

func test(t *testing.T, s Service, workerFailProb, groupCancelProb, taskFailProb float64, msg ...string) {
	numGroups := 10
	numTasks := 10
	numWorkers := 5
	// Set up workers.
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	workerEg, errCtx := errgroup.WithContext(workerCtx)
	for i := 0; i < numWorkers; i++ {
		workerEg.Go(func() error {
			src := s.NewSource("")
			for {
				if err := func() error {
					ctx, cancel := context.WithCancel(errCtx)
					defer cancel()
					err := src.Iterate(ctx, func(_ context.Context, input *types.Any) (*types.Any, error) {
						if rand.Float64() < workerFailProb {
							cancel()
							return nil, nil
						}
						if rand.Float64() < taskFailProb {
							return nil, errTaskFailure
						}
						testTask, err := deserializeTestTask(input)
						if err != nil {
							return nil, err
						}
						return serializeTestTask(testTask)
					})
					if errors.Is(ctx.Err(), context.Canceled) {
						return nil
					}
					return errors.EnsureStack(err)
				}(); err != nil {
					return err
				}
				if errors.Is(workerCtx.Err(), context.Canceled) {
					return nil
				}
			}
		})
	}
	created := [][]bool{}
	collected := [][]bool{}
	for i := 0; i < numGroups; i++ {
		created = append(created, make([]bool, numTasks))
		collected = append(collected, make([]bool, numTasks))
	}
	// Create groups.
	var groupEg errgroup.Group
	for i := 0; i < numGroups; i++ {
		i := i
		groupEg.Go(func() error {
			var inputs []*types.Any
			for j := 0; j < numTasks; j++ {
				input, err := serializeTestTask(&TestTask{ID: strconv.Itoa(j)})
				if err != nil {
					return err
				}
				inputs = append(inputs, input)
				created[i][j] = true
			}
			ctx, cancel := context.WithCancel(errCtx)
			defer cancel()
			d := s.NewDoer("", strconv.Itoa(i), nil)
			if err := DoBatch(ctx, d, inputs, func(j int64, output *types.Any, err error) error {
				if rand.Float64() < groupCancelProb {
					created[i] = nil
					collected[i] = nil
					cancel()
					return nil
				}
				if err != nil {
					if err.Error() != errTaskFailure.Error() {
						return errors.Errorf("task error message (%v) does not equal expected error message (%v)", err.Error(), errTaskFailure.Error())
					}
				} else {
					_, err = deserializeTestTask(output)
					if err != nil {
						return err
					}
				}
				collected[i][j] = true
				return nil
			}); err != nil && !errors.Is(ctx.Err(), context.Canceled) {
				return err
			}
			return nil
		})
	}
	require.NoError(t, groupEg.Wait(), msg)
	workerCancel()
	require.NoError(t, workerEg.Wait(), msg)
	require.Equal(t, created, collected, msg)
}

func TestBasic(t *testing.T) {
	t.Parallel()
	test(t, newTestEtcdService(t), 0, 0, 0, seedRand())
}

func TestWorkerCrashes(t *testing.T) {
	t.Parallel()
	test(t, newTestEtcdService(t), 0.1, 0, 0, seedRand())
}

func TestCancelGroups(t *testing.T) {
	t.Parallel()
	test(t, newTestEtcdService(t), 0, 0.05, 0, seedRand())
}

func TestTaskFailures(t *testing.T) {
	t.Parallel()
	test(t, newTestEtcdService(t), 0, 0, 0.1, seedRand())
}

func TestEverything(t *testing.T) {
	t.Parallel()
	test(t, newTestEtcdService(t), 0.1, 0.2, 0.1, seedRand())
}

func TestRunZeroTasks(t *testing.T) {
	t.Parallel()
	env := testetcd.NewEnv(t)
	s := NewEtcdService(env.EtcdClient, "")
	d := s.NewDoer("", "", nil)
	require.NoError(t, DoBatch(context.Background(), d, nil, func(_ int64, _ *types.Any, _ error) error {
		return errors.New("no tasks should exist")
	}))
}

func TestListTask(t *testing.T) {
	t.Parallel()
	env := testetcd.NewEnv(t)
	testNamespace := tu.UniqueString(t.Name())
	s := NewEtcdService(env.EtcdClient, "")

	numGroups := 10
	numTasks := 10
	numWorkers := 5

	claimedChan := make(chan struct{}, numWorkers)
	finishChan := make(chan struct{}, numWorkers)

	// deterministic failure for easy checking
	shouldFail := func(id string) bool {
		asInt, err := strconv.Atoi(id)
		require.NoError(t, err)
		return asInt%3 == 0
	}

	flushTasksAndVerify := func(flush bool) {
		if flush {
			for i := 0; i < numWorkers; i++ {
				<-claimedChan
			}
		}

		listTask := func(namespace, group string) ([]*taskapi.TaskInfo, error) {
			var out []*taskapi.TaskInfo
			req := &taskapi.ListTaskRequest{Group: &taskapi.Group{
				Namespace: namespace,
				Group:     group,
			}}
			if err := List(context.Background(), s, req, func(info *taskapi.TaskInfo) error {
				out = append(out, info)
				return nil
			}); err != nil {
				return nil, err
			}
			return out, nil
		}

		var groupTotalClaimed, totalClaimed int
		for g := 0; g < numGroups; g++ {
			tasks, err := listTask(testNamespace, strconv.Itoa(g))
			require.NoError(t, err)
			for i, info := range tasks {
				switch info.State {
				case taskapi.State_SUCCESS, taskapi.State_FAILURE:
					// tasks should be most-recently-created first, per group
					order := len(tasks) - 1 - i
					require.Equal(t, info.State == taskapi.State_FAILURE, (g*numTasks+order)%3 == 0)
				case taskapi.State_CLAIMED:
					groupTotalClaimed++
				default:
					require.Equal(t, taskapi.State_RUNNING, info.State)
				}
				asInt, err := strconv.Atoi(info.Group.Group)
				require.NoError(t, err)
				require.Equal(t, asInt, g)
			}
		}
		allTasks, err := listTask(testNamespace, "")
		require.NoError(t, err)

		for _, info := range allTasks {
			if info.State == taskapi.State_CLAIMED {
				totalClaimed++
			}
		}
		// it's possible for per-group results to be deleted between the API calls,
		// but claimed, unfinished tasks must still be present
		require.Equal(t, totalClaimed, groupTotalClaimed)

		if flush {
			require.Equal(t, numWorkers, groupTotalClaimed)
			for i := 0; i < numWorkers; i++ {
				finishChan <- struct{}{}
			}
		} else {
			require.Equal(t, 0, groupTotalClaimed)
		}
	}

	var groupEg errgroup.Group
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	workerEg, errCtx := errgroup.WithContext(workerCtx)
	for g := 0; g < numGroups; g++ {
		g := g
		groupEg.Go(func() error {
			var inputs []*types.Any
			for j := 0; j < numTasks; j++ {
				input, err := serializeTestTask(&TestTask{ID: strconv.Itoa(g*numTasks + j)})
				if err != nil {
					return err
				}
				inputs = append(inputs, input)
			}
			ctx, cancel := context.WithCancel(errCtx)
			defer cancel()
			d := s.NewDoer(testNamespace, strconv.Itoa(g), nil)
			if err := DoBatch(ctx, d, inputs, func(j int64, output *types.Any, err error) error {
				if err != nil {
					if err.Error() != errTaskFailure.Error() {
						return errors.Errorf("task error message (%v) does not equal expected error message (%v)", err.Error(), errTaskFailure.Error())
					}
				} else {
					_, err = deserializeTestTask(output)
					if err != nil {
						return err
					}
				}
				return nil
			}); err != nil && !errors.Is(ctx.Err(), context.Canceled) {
				return err
			}
			return nil
		})
	}

	// check there's no task progress
	flushTasksAndVerify(false)

	for i := 0; i < numWorkers; i++ {
		workerEg.Go(func() error {
			src := s.NewSource("")
			for {
				if err := func() error {
					ctx, cancel := context.WithCancel(errCtx)
					defer cancel()
					err := src.Iterate(ctx, func(_ context.Context, input *types.Any) (*types.Any, error) {
						testTask, err := deserializeTestTask(input)
						if err != nil {
							return nil, err
						}
						// use channels to control task progress
						claimedChan <- struct{}{}
						<-finishChan
						if shouldFail(testTask.ID) {
							return nil, errTaskFailure
						}
						return serializeTestTask(testTask)
					})
					if errors.Is(ctx.Err(), context.Canceled) {
						return nil
					}
					return errors.EnsureStack(err)
				}(); err != nil {
					return err
				}
				if errors.Is(workerCtx.Err(), context.Canceled) {
					return nil
				}
			}
		})
	}

	// allow numWorkers tasks at a time to progress, and check ListTask results
	for i := 0; i < numTasks*numGroups; i += numWorkers {
		flushTasksAndVerify(true)
	}
	close(claimedChan)
	close(finishChan)
	require.NoError(t, groupEg.Wait())
	workerCancel()
	require.NoError(t, workerEg.Wait())
}
