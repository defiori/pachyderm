package cmds

import (
	"fmt"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/server/debug/shell"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing debug commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var duration time.Duration
	var pachd bool
	var database bool
	var pipeline string
	var worker string
	profile := &cobra.Command{
		Use:   "{{alias}} <profile> <file>",
		Short: "Collect a set of pprof profiles.",
		Long:  "Collect a set of pprof profiles.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := client.NewOnUserMachine("debug-profile")
			if err != nil {
				return err
			}
			defer client.Close()
			var d *types.Duration
			if duration != 0 {
				d = types.DurationProto(duration)
			}
			p := &debug.Profile{
				Name:     args[0],
				Duration: d,
			}
			filter, err := createFilter(pachd, database, pipeline, worker)
			if err != nil {
				return err
			}
			return withFile(args[1], func(f *os.File) error {
				return client.Profile(p, filter, f)
			})
		}),
	}
	profile.Flags().DurationVarP(&duration, "duration", "d", time.Minute, "Duration to run a CPU profile for.")
	profile.Flags().BoolVar(&pachd, "pachd", false, "Only collect the profile from pachd.")
	profile.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Only collect the profile from the worker pods for the given pipeline.")
	profile.Flags().StringVarP(&worker, "worker", "w", "", "Only collect the profile from the given worker pod.")
	commands = append(commands, cmdutil.CreateAlias(profile, "debug profile"))

	binary := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Collect a set of binaries.",
		Long:  "Collect a set of binaries.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine("debug-binary")
			if err != nil {
				return err
			}
			defer client.Close()
			filter, err := createFilter(pachd, database, pipeline, worker)
			if err != nil {
				return err
			}
			return withFile(args[0], func(f *os.File) error {
				return client.Binary(filter, f)
			})
		}),
	}
	binary.Flags().BoolVar(&pachd, "pachd", false, "Only collect the binary from pachd.")
	binary.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Only collect the binary from the worker pods for the given pipeline.")
	binary.Flags().StringVarP(&worker, "worker", "w", "", "Only collect the binary from the given worker pod.")
	commands = append(commands, cmdutil.CreateAlias(binary, "debug binary"))

	var limit int64
	dump := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Collect a standard set of debugging information.",
		Long:  "Collect a standard set of debugging information.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine("debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			filter, err := createFilter(pachd, database, pipeline, worker)
			if err != nil {
				return err
			}
			return withFile(args[0], func(f *os.File) error {
				return client.Dump(filter, limit, f)
			})
		}),
	}
	dump.Flags().BoolVar(&pachd, "pachd", false, "Only collect the dump from pachd.")
	dump.Flags().BoolVar(&database, "database", false, "Only collect the dump from pachd's database.")
	dump.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Only collect the dump from the worker pods for the given pipeline.")
	dump.Flags().StringVarP(&worker, "worker", "w", "", "Only collect the dump from the given worker pod.")
	dump.Flags().Int64VarP(&limit, "limit", "l", 0, "Limit sets the limit for the number of commits / jobs that are returned for each repo / pipeline in the dump.")
	commands = append(commands, cmdutil.CreateAlias(dump, "debug dump"))

	var serverPort int
	analyze := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Start a local pachd server to analyze a debug dump.",
		Long:  "Start a local pachd server to analyze a debug dump.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			dump := shell.NewDumpServer(args[0], uint16(serverPort))
			fmt.Println("listening on", dump.Address())
			select {}
		}),
	}
	analyze.Flags().IntVarP(&serverPort, "port", "p", 0,
		"launch a debug server on the given port. If unset, choose a free port automatically")
	commands = append(commands, cmdutil.CreateAlias(analyze, "debug analyze"))

	debug := &cobra.Command{
		Short: "Debug commands for analyzing a running cluster.",
		Long:  "Debug commands for analyzing a running cluster.",
	}
	commands = append(commands, cmdutil.CreateAlias(debug, "debug"))

	return commands
}

// FIXME(CORE-1078): handle projects
func createFilter(pachd, database bool, pipeline, worker string) (*debug.Filter, error) {
	var f *debug.Filter
	if pachd {
		f = &debug.Filter{Filter: &debug.Filter_Pachd{Pachd: true}}
	}
	if database {
		f = &debug.Filter{Filter: &debug.Filter_Database{Database: true}}
	}
	if pipeline != "" {
		if f != nil {
			return nil, errors.Errorf("only one debug filter allowed")
		}
		f = &debug.Filter{Filter: &debug.Filter_Pipeline{Pipeline: &pps.Pipeline{Name: pipeline}}}
	}
	if worker != "" {
		if f != nil {
			return nil, errors.Errorf("only one debug filter allowed")
		}
		f = &debug.Filter{Filter: &debug.Filter_Worker{Worker: &debug.Worker{Pod: worker}}}
	}
	return f, nil
}

func withFile(file string, cb func(*os.File) error) (retErr error) {
	f, err := os.Create(file)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer func() {
		if err := f.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(f)
}
