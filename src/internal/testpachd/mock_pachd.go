package testpachd

import (
	"context"
	"net"

	"github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	authmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	"github.com/pachyderm/pachyderm/v2/src/task"
	"github.com/pachyderm/pachyderm/v2/src/transaction"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	errorsmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/errors"
	loggingmw "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth"
	version "github.com/pachyderm/pachyderm/v2/src/version/versionpb"
)

/* Admin Server Mocks */

type inspectClusterFunc func(context.Context, *types.Empty) (*admin.ClusterInfo, error)

type mockInspectCluster struct{ handler inspectClusterFunc }

func (mock *mockInspectCluster) Use(cb inspectClusterFunc) { mock.handler = cb }

type adminServerAPI struct {
	mock *mockAdminServer
}

type mockAdminServer struct {
	api            adminServerAPI
	InspectCluster mockInspectCluster
}

func (api *adminServerAPI) InspectCluster(ctx context.Context, req *types.Empty) (*admin.ClusterInfo, error) {
	if api.mock.InspectCluster.handler != nil {
		return api.mock.InspectCluster.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock admin.InspectCluster")
}

/* Auth Server Mocks */

type activateAuthFunc func(context.Context, *auth.ActivateRequest) (*auth.ActivateResponse, error)
type deactivateAuthFunc func(context.Context, *auth.DeactivateRequest) (*auth.DeactivateResponse, error)
type getConfigurationFunc func(context.Context, *auth.GetConfigurationRequest) (*auth.GetConfigurationResponse, error)
type setConfigurationFunc func(context.Context, *auth.SetConfigurationRequest) (*auth.SetConfigurationResponse, error)

type modifyRoleBindingFunc func(context.Context, *auth.ModifyRoleBindingRequest) (*auth.ModifyRoleBindingResponse, error)
type getRoleBindingFunc func(context.Context, *auth.GetRoleBindingRequest) (*auth.GetRoleBindingResponse, error)

type authenticateFunc func(context.Context, *auth.AuthenticateRequest) (*auth.AuthenticateResponse, error)
type authorizeFunc func(context.Context, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error)
type getPermissionsFunc func(context.Context, *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error)
type getPermissionsForPrincipalFunc func(context.Context, *auth.GetPermissionsForPrincipalRequest) (*auth.GetPermissionsResponse, error)
type whoAmIFunc func(context.Context, *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error)
type getRolesForPermissionFunc func(context.Context, *auth.GetRolesForPermissionRequest) (*auth.GetRolesForPermissionResponse, error)
type getOIDCLoginFunc func(context.Context, *auth.GetOIDCLoginRequest) (*auth.GetOIDCLoginResponse, error)
type getRobotTokenFunc func(context.Context, *auth.GetRobotTokenRequest) (*auth.GetRobotTokenResponse, error)
type revokeAuthTokenFunc func(context.Context, *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error)
type revokeAuthTokensForUserFunc func(context.Context, *auth.RevokeAuthTokensForUserRequest) (*auth.RevokeAuthTokensForUserResponse, error)
type setGroupsForUserFunc func(context.Context, *auth.SetGroupsForUserRequest) (*auth.SetGroupsForUserResponse, error)
type modifyMembersFunc func(context.Context, *auth.ModifyMembersRequest) (*auth.ModifyMembersResponse, error)
type getGroupsFunc func(context.Context, *auth.GetGroupsRequest) (*auth.GetGroupsResponse, error)
type getGroupsForPrincipalFunc func(context.Context, *auth.GetGroupsForPrincipalRequest) (*auth.GetGroupsResponse, error)
type getUsersFunc func(context.Context, *auth.GetUsersRequest) (*auth.GetUsersResponse, error)
type extractAuthTokensFunc func(context.Context, *auth.ExtractAuthTokensRequest) (*auth.ExtractAuthTokensResponse, error)
type restoreAuthTokenFunc func(context.Context, *auth.RestoreAuthTokenRequest) (*auth.RestoreAuthTokenResponse, error)
type deleteExpiredAuthTokensFunc func(context.Context, *auth.DeleteExpiredAuthTokensRequest) (*auth.DeleteExpiredAuthTokensResponse, error)
type RotateRootTokenFunc func(context.Context, *auth.RotateRootTokenRequest) (*auth.RotateRootTokenResponse, error)

type checkRepoIsAuthorizedFunc func(context.Context, *pfs.Repo, ...auth.Permission) error
type checkClusterIsAuthorizedFunc func(context.Context, ...auth.Permission) error
type checkClusterIsAuthorizedInTransactionFunc func(*txncontext.TransactionContext, ...auth.Permission) error
type checkRepoIsAuthorizedInTransactionFunc func(*txncontext.TransactionContext, *pfs.Repo, ...auth.Permission) error
type authorizeInTransactionFunc func(*txncontext.TransactionContext, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error)
type modifyRoleBindingInTransactionFunc func(*txncontext.TransactionContext, *auth.ModifyRoleBindingRequest) (*auth.ModifyRoleBindingResponse, error)
type getRoleBindingInTransactionFunc func(*txncontext.TransactionContext, *auth.GetRoleBindingRequest) (*auth.GetRoleBindingResponse, error)
type addPipelineReaderToRepoInTransactionFunc func(*txncontext.TransactionContext, string, string) error
type addPipelineWriterToRepoInTransactionFunc func(*txncontext.TransactionContext, string) error
type addPipelineWriterToSourceRepoInTransactionFunc func(*txncontext.TransactionContext, string, string) error
type removePipelineReaderFromRepoInTransactionFunc func(*txncontext.TransactionContext, string, string) error
type createRoleBindingInTransactionFunc func(*txncontext.TransactionContext, string, []string, *auth.Resource) error
type deleteRoleBindingInTransactionFunc func(*txncontext.TransactionContext, *auth.Resource) error
type getPipelineAuthTokenInTransactionFunc func(*txncontext.TransactionContext, string) (string, error)
type revokeAuthTokenInTransactionFunc func(*txncontext.TransactionContext, *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error)
type getPermissionsInTransactionFunc func(*txncontext.TransactionContext, *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error)

type mockActivateAuth struct{ handler activateAuthFunc }
type mockDeactivateAuth struct{ handler deactivateAuthFunc }
type mockGetConfiguration struct{ handler getConfigurationFunc }
type mockSetConfiguration struct{ handler setConfigurationFunc }
type mockModifyRoleBinding struct{ handler modifyRoleBindingFunc }
type mockGetRoleBinding struct{ handler getRoleBindingFunc }

type mockAuthenticate struct{ handler authenticateFunc }
type mockAuthorize struct{ handler authorizeFunc }
type mockGetPermissions struct{ handler getPermissionsFunc }
type mockGetPermissionsForPrincipal struct {
	handler getPermissionsForPrincipalFunc
}
type mockWhoAmI struct{ handler whoAmIFunc }
type mockGetRolesForPermission struct{ handler getRolesForPermissionFunc }
type mockGetOIDCLogin struct{ handler getOIDCLoginFunc }
type mockGetRobotToken struct{ handler getRobotTokenFunc }
type mockRevokeAuthToken struct{ handler revokeAuthTokenFunc }
type mockRevokeAuthTokensForUser struct{ handler revokeAuthTokensForUserFunc }
type mockSetGroupsForUser struct{ handler setGroupsForUserFunc }
type mockModifyMembers struct{ handler modifyMembersFunc }
type mockGetGroups struct{ handler getGroupsFunc }
type mockGetGroupsForPrincipal struct{ handler getGroupsForPrincipalFunc }
type mockGetUsers struct{ handler getUsersFunc }
type mockExtractAuthTokens struct{ handler extractAuthTokensFunc }
type mockRestoreAuthToken struct{ handler restoreAuthTokenFunc }
type mockDeleteExpiredAuthTokens struct{ handler deleteExpiredAuthTokensFunc }
type mockRotateRootToken struct{ handler RotateRootTokenFunc }

type mockCheckRepoIsAuthorized struct {
	handler checkRepoIsAuthorizedFunc
}
type mockCheckClusterIsAuthorized struct {
	handler checkClusterIsAuthorizedFunc
}
type mockCheckClusterIsAuthorizedInTransaction struct {
	handler checkClusterIsAuthorizedInTransactionFunc
}
type mockCheckRepoIsAuthorizedInTransaction struct {
	handler checkRepoIsAuthorizedInTransactionFunc
}
type mockAuthorizeInTransaction struct {
	handler authorizeInTransactionFunc
}
type mockModifyRoleBindingInTransaction struct {
	handler modifyRoleBindingInTransactionFunc
}
type mockGetRoleBindingInTransaction struct {
	handler getRoleBindingInTransactionFunc
}
type mockAddPipelineReaderToRepoInTransaction struct {
	handler addPipelineReaderToRepoInTransactionFunc
}
type mockAddPipelineWriterToRepoInTransaction struct {
	handler addPipelineWriterToRepoInTransactionFunc
}
type mockAddPipelineWriterToSourceRepoInTransaction struct {
	handler addPipelineWriterToSourceRepoInTransactionFunc
}
type mockRemovePipelineReaderFromRepoInTransaction struct {
	handler removePipelineReaderFromRepoInTransactionFunc
}
type mockCreateRoleBindingInTransaction struct {
	handler createRoleBindingInTransactionFunc
}
type mockDeleteRoleBindingInTransaction struct {
	handler deleteRoleBindingInTransactionFunc
}
type mockGetPipelineAuthTokenInTransaction struct {
	handler getPipelineAuthTokenInTransactionFunc
}
type mockRevokeAuthTokenInTransaction struct {
	handler revokeAuthTokenInTransactionFunc
}
type mockGetPermissionsInTransaction struct {
	handler getPermissionsInTransactionFunc
}

func (mock *mockActivateAuth) Use(cb activateAuthFunc)                             { mock.handler = cb }
func (mock *mockDeactivateAuth) Use(cb deactivateAuthFunc)                         { mock.handler = cb }
func (mock *mockGetConfiguration) Use(cb getConfigurationFunc)                     { mock.handler = cb }
func (mock *mockSetConfiguration) Use(cb setConfigurationFunc)                     { mock.handler = cb }
func (mock *mockModifyRoleBinding) Use(cb modifyRoleBindingFunc)                   { mock.handler = cb }
func (mock *mockGetRoleBinding) Use(cb getRoleBindingFunc)                         { mock.handler = cb }
func (mock *mockAuthenticate) Use(cb authenticateFunc)                             { mock.handler = cb }
func (mock *mockAuthorize) Use(cb authorizeFunc)                                   { mock.handler = cb }
func (mock *mockWhoAmI) Use(cb whoAmIFunc)                                         { mock.handler = cb }
func (mock *mockGetRolesForPermission) Use(cb getRolesForPermissionFunc)           { mock.handler = cb }
func (mock *mockGetOIDCLogin) Use(cb getOIDCLoginFunc)                             { mock.handler = cb }
func (mock *mockGetRobotToken) Use(cb getRobotTokenFunc)                           { mock.handler = cb }
func (mock *mockRevokeAuthToken) Use(cb revokeAuthTokenFunc)                       { mock.handler = cb }
func (mock *mockRevokeAuthTokensForUser) Use(cb revokeAuthTokensForUserFunc)       { mock.handler = cb }
func (mock *mockSetGroupsForUser) Use(cb setGroupsForUserFunc)                     { mock.handler = cb }
func (mock *mockModifyMembers) Use(cb modifyMembersFunc)                           { mock.handler = cb }
func (mock *mockGetGroups) Use(cb getGroupsFunc)                                   { mock.handler = cb }
func (mock *mockGetGroupsForPrincipal) Use(cb getGroupsForPrincipalFunc)           { mock.handler = cb }
func (mock *mockGetPermissions) Use(cb getPermissionsFunc)                         { mock.handler = cb }
func (mock *mockGetPermissionsForPrincipal) Use(cb getPermissionsForPrincipalFunc) { mock.handler = cb }
func (mock *mockGetUsers) Use(cb getUsersFunc)                                     { mock.handler = cb }
func (mock *mockExtractAuthTokens) Use(cb extractAuthTokensFunc)                   { mock.handler = cb }
func (mock *mockRestoreAuthToken) Use(cb restoreAuthTokenFunc)                     { mock.handler = cb }
func (mock *mockDeleteExpiredAuthTokens) Use(cb deleteExpiredAuthTokensFunc)       { mock.handler = cb }
func (mock *mockRotateRootToken) Use(cb RotateRootTokenFunc)                       { mock.handler = cb }

func (mock *mockCheckRepoIsAuthorized) Use(cb checkRepoIsAuthorizedFunc) {
	mock.handler = cb
}
func (mock *mockCheckClusterIsAuthorized) Use(cb checkClusterIsAuthorizedFunc) {
	mock.handler = cb
}
func (mock *mockCheckClusterIsAuthorizedInTransaction) Use(cb checkClusterIsAuthorizedInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockCheckRepoIsAuthorizedInTransaction) Use(cb checkRepoIsAuthorizedInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockAuthorizeInTransaction) Use(cb authorizeInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockModifyRoleBindingInTransaction) Use(cb modifyRoleBindingInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockGetRoleBindingInTransaction) Use(cb getRoleBindingInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockAddPipelineReaderToRepoInTransaction) Use(cb addPipelineReaderToRepoInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockAddPipelineWriterToRepoInTransaction) Use(cb addPipelineWriterToRepoInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockAddPipelineWriterToSourceRepoInTransaction) Use(cb addPipelineWriterToSourceRepoInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockRemovePipelineReaderFromRepoInTransaction) Use(cb removePipelineReaderFromRepoInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockCreateRoleBindingInTransaction) Use(cb createRoleBindingInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockDeleteRoleBindingInTransaction) Use(cb deleteRoleBindingInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockGetPipelineAuthTokenInTransaction) Use(cb getPipelineAuthTokenInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockRevokeAuthTokenInTransaction) Use(cb revokeAuthTokenInTransactionFunc) {
	mock.handler = cb
}
func (mock *mockGetPermissionsInTransaction) Use(cb getPermissionsInTransactionFunc) {
	mock.handler = cb
}

type authServerAPI struct {
	mock *mockAuthServer
}

type mockAuthServer struct {
	api                                        authServerAPI
	Activate                                   mockActivateAuth
	Deactivate                                 mockDeactivateAuth
	GetConfiguration                           mockGetConfiguration
	SetConfiguration                           mockSetConfiguration
	ModifyRoleBinding                          mockModifyRoleBinding
	GetRoleBinding                             mockGetRoleBinding
	Authenticate                               mockAuthenticate
	Authorize                                  mockAuthorize
	GetPermissions                             mockGetPermissions
	GetPermissionsForPrincipal                 mockGetPermissionsForPrincipal
	WhoAmI                                     mockWhoAmI
	GetRolesForPermission                      mockGetRolesForPermission
	GetOIDCLogin                               mockGetOIDCLogin
	GetRobotToken                              mockGetRobotToken
	RevokeAuthToken                            mockRevokeAuthToken
	RevokeAuthTokensForUser                    mockRevokeAuthTokensForUser
	SetGroupsForUser                           mockSetGroupsForUser
	ModifyMembers                              mockModifyMembers
	GetGroups                                  mockGetGroups
	GetGroupsForPrincipal                      mockGetGroupsForPrincipal
	GetUsers                                   mockGetUsers
	ExtractAuthTokens                          mockExtractAuthTokens
	RestoreAuthToken                           mockRestoreAuthToken
	DeleteExpiredAuthTokens                    mockDeleteExpiredAuthTokens
	RotateRootToken                            mockRotateRootToken
	CheckRepoIsAuthorized                      mockCheckRepoIsAuthorized
	CheckClusterIsAuthorized                   mockCheckClusterIsAuthorized
	CheckClusterIsAuthorizedInTransaction      mockCheckClusterIsAuthorizedInTransaction
	CheckRepoIsAuthorizedInTransaction         mockCheckRepoIsAuthorizedInTransaction
	AuthorizeInTransaction                     mockAuthorizeInTransaction
	ModifyRoleBindingInTransaction             mockModifyRoleBindingInTransaction
	GetRoleBindingInTransaction                mockGetRoleBindingInTransaction
	AddPipelineReaderToRepoInTransaction       mockAddPipelineReaderToRepoInTransaction
	AddPipelineWriterToRepoInTransaction       mockAddPipelineWriterToRepoInTransaction
	AddPipelineWriterToSourceRepoInTransaction mockAddPipelineWriterToSourceRepoInTransaction
	RemovePipelineReaderFromRepoInTransaction  mockRemovePipelineReaderFromRepoInTransaction
	CreateRoleBindingInTransaction             mockCreateRoleBindingInTransaction
	DeleteRoleBindingInTransaction             mockDeleteRoleBindingInTransaction
	GetPipelineAuthTokenInTransaction          mockGetPipelineAuthTokenInTransaction
	RevokeAuthTokenInTransaction               mockRevokeAuthTokenInTransaction
	GetPermissionsInTransaction                mockGetPermissionsInTransaction
}

func (api *authServerAPI) Activate(ctx context.Context, req *auth.ActivateRequest) (*auth.ActivateResponse, error) {
	if api.mock.Activate.handler != nil {
		return api.mock.Activate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.Activate")
}
func (api *authServerAPI) Deactivate(ctx context.Context, req *auth.DeactivateRequest) (*auth.DeactivateResponse, error) {
	if api.mock.Deactivate.handler != nil {
		return api.mock.Deactivate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.Deactivate")
}
func (api *authServerAPI) GetConfiguration(ctx context.Context, req *auth.GetConfigurationRequest) (*auth.GetConfigurationResponse, error) {
	if api.mock.GetConfiguration.handler != nil {
		return api.mock.GetConfiguration.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetConfiguration")
}
func (api *authServerAPI) SetConfiguration(ctx context.Context, req *auth.SetConfigurationRequest) (*auth.SetConfigurationResponse, error) {
	if api.mock.SetConfiguration.handler != nil {
		return api.mock.SetConfiguration.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.SetConfiguration")
}
func (api *authServerAPI) GetRoleBinding(ctx context.Context, req *auth.GetRoleBindingRequest) (*auth.GetRoleBindingResponse, error) {
	if api.mock.GetRoleBinding.handler != nil {
		return api.mock.GetRoleBinding.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetRoleBinding")
}
func (api *authServerAPI) ModifyRoleBinding(ctx context.Context, req *auth.ModifyRoleBindingRequest) (*auth.ModifyRoleBindingResponse, error) {
	if api.mock.ModifyRoleBinding.handler != nil {
		return api.mock.ModifyRoleBinding.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.ModifyRoleBinding")
}
func (api *authServerAPI) Authenticate(ctx context.Context, req *auth.AuthenticateRequest) (*auth.AuthenticateResponse, error) {
	if api.mock.Authenticate.handler != nil {
		return api.mock.Authenticate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.Authenticate")
}
func (api *authServerAPI) GetPermissions(ctx context.Context, req *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error) {
	if api.mock.GetPermissions.handler != nil {
		return api.mock.GetPermissions.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetPermissions")
}
func (api *authServerAPI) GetPermissionsForPrincipal(ctx context.Context, req *auth.GetPermissionsForPrincipalRequest) (*auth.GetPermissionsResponse, error) {
	if api.mock.GetPermissionsForPrincipal.handler != nil {
		return api.mock.GetPermissionsForPrincipal.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetPermissions")
}
func (api *authServerAPI) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	if api.mock.Authorize.handler != nil {
		return api.mock.Authorize.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.Authorize")
}
func (api *authServerAPI) WhoAmI(ctx context.Context, req *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error) {
	if api.mock.WhoAmI.handler != nil {
		return api.mock.WhoAmI.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.WhoAmI")
}
func (api *authServerAPI) GetRolesForPermission(ctx context.Context, req *auth.GetRolesForPermissionRequest) (*auth.GetRolesForPermissionResponse, error) {
	if api.mock.GetRolesForPermission.handler != nil {
		return api.mock.GetRolesForPermission.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetRolesForPermission")
}
func (api *authServerAPI) GetOIDCLogin(ctx context.Context, req *auth.GetOIDCLoginRequest) (*auth.GetOIDCLoginResponse, error) {
	if api.mock.GetOIDCLogin.handler != nil {
		return api.mock.GetOIDCLogin.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetOIDCLogin")
}
func (api *authServerAPI) GetRobotToken(ctx context.Context, req *auth.GetRobotTokenRequest) (*auth.GetRobotTokenResponse, error) {
	if api.mock.GetRobotToken.handler != nil {
		return api.mock.GetRobotToken.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetRobotToken")
}
func (api *authServerAPI) RevokeAuthToken(ctx context.Context, req *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error) {
	if api.mock.RevokeAuthToken.handler != nil {
		return api.mock.RevokeAuthToken.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.RevokeAuthToken")
}
func (api *authServerAPI) RevokeAuthTokensForUser(ctx context.Context, req *auth.RevokeAuthTokensForUserRequest) (*auth.RevokeAuthTokensForUserResponse, error) {
	if api.mock.RevokeAuthTokensForUser.handler != nil {
		return api.mock.RevokeAuthTokensForUser.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.RevokeAuthTokensForUser")
}
func (api *authServerAPI) SetGroupsForUser(ctx context.Context, req *auth.SetGroupsForUserRequest) (*auth.SetGroupsForUserResponse, error) {
	if api.mock.SetGroupsForUser.handler != nil {
		return api.mock.SetGroupsForUser.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.SetGroupsForUser")
}
func (api *authServerAPI) ModifyMembers(ctx context.Context, req *auth.ModifyMembersRequest) (*auth.ModifyMembersResponse, error) {
	if api.mock.ModifyMembers.handler != nil {
		return api.mock.ModifyMembers.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.ModifyMembers")
}
func (api *authServerAPI) GetGroups(ctx context.Context, req *auth.GetGroupsRequest) (*auth.GetGroupsResponse, error) {
	if api.mock.GetGroups.handler != nil {
		return api.mock.GetGroups.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetGroups")
}
func (api *authServerAPI) GetGroupsForPrincipal(ctx context.Context, req *auth.GetGroupsForPrincipalRequest) (*auth.GetGroupsResponse, error) {
	if api.mock.GetGroupsForPrincipal.handler != nil {
		return api.mock.GetGroupsForPrincipal.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetGroupsForPrincipal")
}
func (api *authServerAPI) GetUsers(ctx context.Context, req *auth.GetUsersRequest) (*auth.GetUsersResponse, error) {
	if api.mock.GetUsers.handler != nil {
		return api.mock.GetUsers.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetUsers")
}

func (api *authServerAPI) ExtractAuthTokens(ctx context.Context, req *auth.ExtractAuthTokensRequest) (*auth.ExtractAuthTokensResponse, error) {
	if api.mock.ExtractAuthTokens.handler != nil {
		return api.mock.ExtractAuthTokens.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.ExtractAuthTokens")
}

func (api *authServerAPI) RestoreAuthToken(ctx context.Context, req *auth.RestoreAuthTokenRequest) (*auth.RestoreAuthTokenResponse, error) {
	if api.mock.RestoreAuthToken.handler != nil {
		return api.mock.RestoreAuthToken.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.RestoreAuthToken")
}

func (api *authServerAPI) DeleteExpiredAuthTokens(ctx context.Context, req *auth.DeleteExpiredAuthTokensRequest) (*auth.DeleteExpiredAuthTokensResponse, error) {
	if api.mock.DeleteExpiredAuthTokens.handler != nil {
		return api.mock.DeleteExpiredAuthTokens.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.DeleteExpiredAuthTokens")
}

func (api *authServerAPI) RotateRootToken(ctx context.Context, req *auth.RotateRootTokenRequest) (*auth.RotateRootTokenResponse, error) {
	if api.mock.RotateRootToken.handler != nil {
		return api.mock.RotateRootToken.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.RotateRootToken")
}

func (api *authServerAPI) CheckRepoIsAuthorized(ctx context.Context, repo *pfs.Repo, permission ...auth.Permission) error {
	if api.mock.CheckRepoIsAuthorized.handler != nil {
		return api.mock.CheckRepoIsAuthorized.handler(ctx, repo, permission...)
	}
	return errors.Errorf("unhandled pachd mock auth.CheckRepoIsAuthorized")
}

func (api *authServerAPI) CheckClusterIsAuthorized(ctx context.Context, p ...auth.Permission) error {
	if api.mock.CheckClusterIsAuthorized.handler != nil {
		return api.mock.CheckClusterIsAuthorized.handler(ctx, p...)
	}
	return errors.Errorf("unhandled pachd mock auth.CheckClusterIsAuthorized")
}

func (api *authServerAPI) CheckClusterIsAuthorizedInTransaction(transactionContext *txncontext.TransactionContext, permission ...auth.Permission) error {
	if api.mock.CheckClusterIsAuthorizedInTransaction.handler != nil {
		return api.mock.CheckClusterIsAuthorizedInTransaction.handler(transactionContext, permission...)
	}
	return errors.Errorf("unhandled pachd mock auth.CheckClusterIsAuthorizedInTranscation")
}

func (api *authServerAPI) CheckRepoIsAuthorizedInTransaction(transactionContext *txncontext.TransactionContext, repo *pfs.Repo, permission ...auth.Permission) error {
	if api.mock.CheckRepoIsAuthorizedInTransaction.handler != nil {
		return api.mock.CheckRepoIsAuthorizedInTransaction.handler(transactionContext, repo, permission...)
	}
	return errors.Errorf("unhandled pachd mock auth.CheckRepoIsAuthorizedInTranscation")
}

func (api *authServerAPI) AuthorizeInTransaction(transactionContext *txncontext.TransactionContext, request *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	if api.mock.AuthorizeInTransaction.handler != nil {
		return api.mock.AuthorizeInTransaction.handler(transactionContext, request)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.AuthorizeInTransaction")
}

func (api *authServerAPI) ModifyRoleBindingInTransaction(transactionContext *txncontext.TransactionContext, request *auth.ModifyRoleBindingRequest) (*auth.ModifyRoleBindingResponse, error) {
	if api.mock.ModifyRoleBindingInTransaction.handler != nil {
		return api.mock.ModifyRoleBindingInTransaction.handler(transactionContext, request)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.ModifyRoleBindingInTransaction")
}

func (api *authServerAPI) GetRoleBindingInTransaction(transactionContext *txncontext.TransactionContext, request *auth.GetRoleBindingRequest) (*auth.GetRoleBindingResponse, error) {
	if api.mock.GetRoleBindingInTransaction.handler != nil {
		return api.mock.GetRoleBindingInTransaction.handler(transactionContext, request)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetRoleBindingInTransaction")
}

func (api *authServerAPI) AddPipelineReaderToRepoInTransaction(transactionContext *txncontext.TransactionContext, s string, s2 string) error {
	if api.mock.AddPipelineReaderToRepoInTransaction.handler != nil {
		return api.mock.AddPipelineReaderToRepoInTransaction.handler(transactionContext, s, s2)
	}
	return errors.Errorf("unhandled pachd mock auth.AddPipelineReaderToRepoInTransaction")
}

func (api *authServerAPI) AddPipelineWriterToRepoInTransaction(transactionContext *txncontext.TransactionContext, s string) error {
	if api.mock.AddPipelineWriterToRepoInTransaction.handler != nil {
		return api.mock.AddPipelineWriterToRepoInTransaction.handler(transactionContext, s)
	}
	return errors.Errorf("unhandled pachd mock auth.AddPipelineWriterToRepoInTransaction")
}

func (api *authServerAPI) AddPipelineWriterToSourceRepoInTransaction(transactionContext *txncontext.TransactionContext, s string, s2 string) error {
	if api.mock.AddPipelineWriterToSourceRepoInTransaction.handler != nil {
		return api.mock.AddPipelineWriterToSourceRepoInTransaction.handler(transactionContext, s, s2)
	}
	return errors.Errorf("unhandled pachd mock auth.AddPipelineWriterToSourceRepoInTransaction")
}

func (api *authServerAPI) RemovePipelineReaderFromRepoInTransaction(transactionContext *txncontext.TransactionContext, s string, s2 string) error {
	if api.mock.RemovePipelineReaderFromRepoInTransaction.handler != nil {
		return api.mock.RemovePipelineReaderFromRepoInTransaction.handler(transactionContext, s, s2)
	}
	return errors.Errorf("unhandled pachd mock auth.RemovePipelineReaderFromRepoInTransaction")
}

func (api *authServerAPI) CreateRoleBindingInTransaction(transactionContext *txncontext.TransactionContext, s string, strings []string, resource *auth.Resource) error {
	if api.mock.CreateRoleBindingInTransaction.handler != nil {
		return api.mock.CreateRoleBindingInTransaction.handler(transactionContext, s, strings, resource)
	}
	return errors.Errorf("unhandled pachd mock auth.CreateRoleBindingInTransaction")
}

func (api *authServerAPI) DeleteRoleBindingInTransaction(transactionContext *txncontext.TransactionContext, resource *auth.Resource) error {
	if api.mock.DeleteRoleBindingInTransaction.handler != nil {
		return api.mock.DeleteRoleBindingInTransaction.handler(transactionContext, resource)
	}
	return errors.Errorf("unhandled pachd mock auth.DeleteRoleBindingInTransaction")
}

func (api *authServerAPI) GetPipelineAuthTokenInTransaction(transactionContext *txncontext.TransactionContext, s string) (string, error) {
	if api.mock.GetPipelineAuthTokenInTransaction.handler != nil {
		return api.mock.GetPipelineAuthTokenInTransaction.handler(transactionContext, s)
	}
	return "", errors.Errorf("unhandled pachd mock auth.GetPipelineAuthTokenInTransaction")
}

func (api *authServerAPI) RevokeAuthTokenInTransaction(transactionContext *txncontext.TransactionContext, request *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error) {
	if api.mock.RevokeAuthTokenInTransaction.handler != nil {
		return api.mock.RevokeAuthTokenInTransaction.handler(transactionContext, request)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.RevokeAuthTokenInTransaction")
}

func (api *authServerAPI) GetPermissionsInTransaction(transactionContext *txncontext.TransactionContext, request *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error) {
	if api.mock.GetPermissionsInTransaction.handler != nil {
		return api.mock.GetPermissionsInTransaction.handler(transactionContext, request)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetPermissionsInTransaction")
}

/* License Server Mocks */
type activateLicenseFunc func(context.Context, *license.ActivateRequest) (*license.ActivateResponse, error)
type getActivationCodeLicenseFunc func(context.Context, *license.GetActivationCodeRequest) (*license.GetActivationCodeResponse, error)
type deleteAllLicenseFunc func(context.Context, *license.DeleteAllRequest) (*license.DeleteAllResponse, error)
type addClusterFunc func(context.Context, *license.AddClusterRequest) (*license.AddClusterResponse, error)
type deleteClusterFunc func(context.Context, *license.DeleteClusterRequest) (*license.DeleteClusterResponse, error)
type listClustersFunc func(context.Context, *license.ListClustersRequest) (*license.ListClustersResponse, error)
type updateClusterFunc func(context.Context, *license.UpdateClusterRequest) (*license.UpdateClusterResponse, error)
type heartbeatLicenseFunc func(context.Context, *license.HeartbeatRequest) (*license.HeartbeatResponse, error)
type listUserClustersFunc func(context.Context, *license.ListUserClustersRequest) (*license.ListUserClustersResponse, error)

type mockActivateLicense struct{ handler activateLicenseFunc }
type mockGetActivationCodeLicense struct{ handler getActivationCodeLicenseFunc }
type mockDeleteAllLicense struct{ handler deleteAllLicenseFunc }
type mockAddCluster struct{ handler addClusterFunc }
type mockDeleteCluster struct{ handler deleteClusterFunc }
type mockListClusters struct{ handler listClustersFunc }
type mockUpdateCluster struct{ handler updateClusterFunc }
type mockHeartbeatLicense struct{ handler heartbeatLicenseFunc }
type mockListUserClusters struct{ handler listUserClustersFunc }

func (mock *mockActivateLicense) Use(cb activateLicenseFunc)                   { mock.handler = cb }
func (mock *mockGetActivationCodeLicense) Use(cb getActivationCodeLicenseFunc) { mock.handler = cb }
func (mock *mockDeleteAllLicense) Use(cb deleteAllLicenseFunc)                 { mock.handler = cb }
func (mock *mockAddCluster) Use(cb addClusterFunc)                             { mock.handler = cb }
func (mock *mockDeleteCluster) Use(cb deleteClusterFunc)                       { mock.handler = cb }
func (mock *mockListClusters) Use(cb listClustersFunc)                         { mock.handler = cb }
func (mock *mockUpdateCluster) Use(cb updateClusterFunc)                       { mock.handler = cb }
func (mock *mockHeartbeatLicense) Use(cb heartbeatLicenseFunc)                 { mock.handler = cb }
func (mock *mockListUserClusters) Use(cb listUserClustersFunc)                 { mock.handler = cb }

type licenseServerAPI struct {
	mock *mockLicenseServer
}

type mockLicenseServer struct {
	api               licenseServerAPI
	Activate          mockActivateLicense
	GetActivationCode mockGetActivationCodeLicense
	DeleteAll         mockDeleteAllLicense
	AddCluster        mockAddCluster
	DeleteCluster     mockDeleteCluster
	ListClusters      mockListClusters
	UpdateCluster     mockUpdateCluster
	Heartbeat         mockHeartbeatLicense
	ListUserClusters  mockListUserClusters
}

func (api *licenseServerAPI) Activate(ctx context.Context, req *license.ActivateRequest) (*license.ActivateResponse, error) {
	if api.mock.Activate.handler != nil {
		return api.mock.Activate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock license.Activate")
}

func (api *licenseServerAPI) GetActivationCode(ctx context.Context, req *license.GetActivationCodeRequest) (*license.GetActivationCodeResponse, error) {
	if api.mock.GetActivationCode.handler != nil {
		return api.mock.GetActivationCode.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock license.GetActivationCode")
}

func (api *licenseServerAPI) DeleteAll(ctx context.Context, req *license.DeleteAllRequest) (*license.DeleteAllResponse, error) {
	if api.mock.DeleteAll.handler != nil {
		return api.mock.DeleteAll.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock license.Activate")
}

func (api *licenseServerAPI) AddCluster(ctx context.Context, req *license.AddClusterRequest) (*license.AddClusterResponse, error) {
	if api.mock.AddCluster.handler != nil {
		return api.mock.AddCluster.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock license.AddCluster")
}

func (api *licenseServerAPI) DeleteCluster(ctx context.Context, req *license.DeleteClusterRequest) (*license.DeleteClusterResponse, error) {
	if api.mock.DeleteCluster.handler != nil {
		return api.mock.DeleteCluster.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock license.DeleteCluster")
}

func (api *licenseServerAPI) ListClusters(ctx context.Context, req *license.ListClustersRequest) (*license.ListClustersResponse, error) {
	if api.mock.ListClusters.handler != nil {
		return api.mock.ListClusters.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock license.ListClusters")
}

func (api *licenseServerAPI) UpdateCluster(ctx context.Context, req *license.UpdateClusterRequest) (*license.UpdateClusterResponse, error) {
	if api.mock.UpdateCluster.handler != nil {
		return api.mock.UpdateCluster.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock license.UpdateCluster")
}

func (api *licenseServerAPI) Heartbeat(ctx context.Context, req *license.HeartbeatRequest) (*license.HeartbeatResponse, error) {
	if api.mock.Heartbeat.handler != nil {
		return api.mock.Heartbeat.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock license.Heartbeat")
}

func (api *licenseServerAPI) ListUserClusters(ctx context.Context, req *license.ListUserClustersRequest) (*license.ListUserClustersResponse, error) {
	if api.mock.ListUserClusters.handler != nil {
		return api.mock.ListUserClusters.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock license.ListUserClusters")
}

/* Identity Server Mocks */
type setIdentityServerConfigFunc func(context.Context, *identity.SetIdentityServerConfigRequest) (*identity.SetIdentityServerConfigResponse, error)
type getIdentityServerConfigFunc func(context.Context, *identity.GetIdentityServerConfigRequest) (*identity.GetIdentityServerConfigResponse, error)
type createIDPConnectorFunc func(context.Context, *identity.CreateIDPConnectorRequest) (*identity.CreateIDPConnectorResponse, error)
type updateIDPConnectorFunc func(context.Context, *identity.UpdateIDPConnectorRequest) (*identity.UpdateIDPConnectorResponse, error)
type listIDPConnectorsFunc func(context.Context, *identity.ListIDPConnectorsRequest) (*identity.ListIDPConnectorsResponse, error)
type getIDPConnectorFunc func(context.Context, *identity.GetIDPConnectorRequest) (*identity.GetIDPConnectorResponse, error)
type deleteIDPConnectorFunc func(context.Context, *identity.DeleteIDPConnectorRequest) (*identity.DeleteIDPConnectorResponse, error)
type createOIDCClientFunc func(context.Context, *identity.CreateOIDCClientRequest) (*identity.CreateOIDCClientResponse, error)
type updateOIDCClientFunc func(context.Context, *identity.UpdateOIDCClientRequest) (*identity.UpdateOIDCClientResponse, error)
type getOIDCClientFunc func(context.Context, *identity.GetOIDCClientRequest) (*identity.GetOIDCClientResponse, error)
type listOIDCClientsFunc func(context.Context, *identity.ListOIDCClientsRequest) (*identity.ListOIDCClientsResponse, error)
type deleteOIDCClientFunc func(context.Context, *identity.DeleteOIDCClientRequest) (*identity.DeleteOIDCClientResponse, error)
type deleteAllFunc func(context.Context, *identity.DeleteAllRequest) (*identity.DeleteAllResponse, error)

type mockSetIdentityServerConfig struct{ handler setIdentityServerConfigFunc }
type mockGetIdentityServerConfig struct{ handler getIdentityServerConfigFunc }
type mockCreateIDPConnector struct{ handler createIDPConnectorFunc }
type mockUpdateIDPConnector struct{ handler updateIDPConnectorFunc }
type mockListIDPConnectors struct{ handler listIDPConnectorsFunc }
type mockGetIDPConnector struct{ handler getIDPConnectorFunc }
type mockDeleteIDPConnector struct{ handler deleteIDPConnectorFunc }
type mockCreateOIDCClient struct{ handler createOIDCClientFunc }
type mockUpdateOIDCClient struct{ handler updateOIDCClientFunc }
type mockGetOIDCClient struct{ handler getOIDCClientFunc }
type mockListOIDCClients struct{ handler listOIDCClientsFunc }
type mockDeleteOIDCClient struct{ handler deleteOIDCClientFunc }
type mockDeleteAll struct{ handler deleteAllFunc }

func (mock *mockSetIdentityServerConfig) Use(cb setIdentityServerConfigFunc) { mock.handler = cb }
func (mock *mockGetIdentityServerConfig) Use(cb getIdentityServerConfigFunc) { mock.handler = cb }
func (mock *mockCreateIDPConnector) Use(cb createIDPConnectorFunc)           { mock.handler = cb }
func (mock *mockUpdateIDPConnector) Use(cb updateIDPConnectorFunc)           { mock.handler = cb }
func (mock *mockListIDPConnectors) Use(cb listIDPConnectorsFunc)             { mock.handler = cb }
func (mock *mockGetIDPConnector) Use(cb getIDPConnectorFunc)                 { mock.handler = cb }
func (mock *mockDeleteIDPConnector) Use(cb deleteIDPConnectorFunc)           { mock.handler = cb }
func (mock *mockCreateOIDCClient) Use(cb createOIDCClientFunc)               { mock.handler = cb }
func (mock *mockUpdateOIDCClient) Use(cb updateOIDCClientFunc)               { mock.handler = cb }
func (mock *mockGetOIDCClient) Use(cb getOIDCClientFunc)                     { mock.handler = cb }
func (mock *mockListOIDCClients) Use(cb listOIDCClientsFunc)                 { mock.handler = cb }
func (mock *mockDeleteOIDCClient) Use(cb deleteOIDCClientFunc)               { mock.handler = cb }
func (mock *mockDeleteAll) Use(cb deleteAllFunc)                             { mock.handler = cb }

type identityServerAPI struct {
	mock *mockIdentityServer
}

type mockIdentityServer struct {
	api                     identityServerAPI
	SetIdentityServerConfig mockSetIdentityServerConfig
	GetIdentityServerConfig mockGetIdentityServerConfig
	CreateIDPConnector      mockCreateIDPConnector
	UpdateIDPConnector      mockUpdateIDPConnector
	ListIDPConnectors       mockListIDPConnectors
	GetIDPConnector         mockGetIDPConnector
	DeleteIDPConnector      mockDeleteIDPConnector
	CreateOIDCClient        mockCreateOIDCClient
	UpdateOIDCClient        mockUpdateOIDCClient
	GetOIDCClient           mockGetOIDCClient
	ListOIDCClients         mockListOIDCClients
	DeleteOIDCClient        mockDeleteOIDCClient
	DeleteAll               mockDeleteAll
}

func (api *identityServerAPI) SetIdentityServerConfig(ctx context.Context, req *identity.SetIdentityServerConfigRequest) (*identity.SetIdentityServerConfigResponse, error) {
	if api.mock.SetIdentityServerConfig.handler != nil {
		return api.mock.SetIdentityServerConfig.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.SetIdentityServerConfig")
}

func (api *identityServerAPI) GetIdentityServerConfig(ctx context.Context, req *identity.GetIdentityServerConfigRequest) (*identity.GetIdentityServerConfigResponse, error) {
	if api.mock.GetIdentityServerConfig.handler != nil {
		return api.mock.GetIdentityServerConfig.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.GetIdentityServerConfig")
}
func (api *identityServerAPI) CreateIDPConnector(ctx context.Context, req *identity.CreateIDPConnectorRequest) (*identity.CreateIDPConnectorResponse, error) {
	if api.mock.CreateIDPConnector.handler != nil {
		return api.mock.CreateIDPConnector.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.CreateIDPConnector")
}
func (api *identityServerAPI) UpdateIDPConnector(ctx context.Context, req *identity.UpdateIDPConnectorRequest) (*identity.UpdateIDPConnectorResponse, error) {
	if api.mock.UpdateIDPConnector.handler != nil {
		return api.mock.UpdateIDPConnector.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.UpdateIDPConnector")
}
func (api *identityServerAPI) ListIDPConnectors(ctx context.Context, req *identity.ListIDPConnectorsRequest) (*identity.ListIDPConnectorsResponse, error) {
	if api.mock.ListIDPConnectors.handler != nil {
		return api.mock.ListIDPConnectors.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.ListIDPConnectors")
}
func (api *identityServerAPI) GetIDPConnector(ctx context.Context, req *identity.GetIDPConnectorRequest) (*identity.GetIDPConnectorResponse, error) {
	if api.mock.GetIDPConnector.handler != nil {
		return api.mock.GetIDPConnector.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.GetIDPConnector")
}
func (api *identityServerAPI) DeleteIDPConnector(ctx context.Context, req *identity.DeleteIDPConnectorRequest) (*identity.DeleteIDPConnectorResponse, error) {
	if api.mock.DeleteIDPConnector.handler != nil {
		return api.mock.DeleteIDPConnector.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.DeleteIDPConnector")
}
func (api *identityServerAPI) CreateOIDCClient(ctx context.Context, req *identity.CreateOIDCClientRequest) (*identity.CreateOIDCClientResponse, error) {
	if api.mock.CreateOIDCClient.handler != nil {
		return api.mock.CreateOIDCClient.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.CreateOIDCClient")
}
func (api *identityServerAPI) UpdateOIDCClient(ctx context.Context, req *identity.UpdateOIDCClientRequest) (*identity.UpdateOIDCClientResponse, error) {
	if api.mock.UpdateOIDCClient.handler != nil {
		return api.mock.UpdateOIDCClient.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.UpdateOIDCClient")
}
func (api *identityServerAPI) GetOIDCClient(ctx context.Context, req *identity.GetOIDCClientRequest) (*identity.GetOIDCClientResponse, error) {
	if api.mock.GetOIDCClient.handler != nil {
		return api.mock.GetOIDCClient.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.identityServerAPI")
}
func (api *identityServerAPI) ListOIDCClients(ctx context.Context, req *identity.ListOIDCClientsRequest) (*identity.ListOIDCClientsResponse, error) {
	if api.mock.ListOIDCClients.handler != nil {
		return api.mock.ListOIDCClients.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.ListOIDCClients")
}
func (api *identityServerAPI) DeleteOIDCClient(ctx context.Context, req *identity.DeleteOIDCClientRequest) (*identity.DeleteOIDCClientResponse, error) {
	if api.mock.DeleteOIDCClient.handler != nil {
		return api.mock.DeleteOIDCClient.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.DeleteOIDCClient")
}
func (api *identityServerAPI) DeleteAll(ctx context.Context, req *identity.DeleteAllRequest) (*identity.DeleteAllResponse, error) {
	if api.mock.DeleteAll.handler != nil {
		return api.mock.DeleteAll.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock identity.DeleteAll")
}

/* Enterprise Server Mocks */

type activateEnterpriseFunc func(context.Context, *enterprise.ActivateRequest) (*enterprise.ActivateResponse, error)
type getStateFunc func(context.Context, *enterprise.GetStateRequest) (*enterprise.GetStateResponse, error)
type getActivationCodeFunc func(context.Context, *enterprise.GetActivationCodeRequest) (*enterprise.GetActivationCodeResponse, error)
type deactivateEnterpriseFunc func(context.Context, *enterprise.DeactivateRequest) (*enterprise.DeactivateResponse, error)
type heartbeatEnterpriseFunc func(context.Context, *enterprise.HeartbeatRequest) (*enterprise.HeartbeatResponse, error)
type pauseFunc func(context.Context, *enterprise.PauseRequest) (*enterprise.PauseResponse, error)
type pauseStatusFunc func(context.Context, *enterprise.PauseStatusRequest) (*enterprise.PauseStatusResponse, error)
type unpauseFunc func(context.Context, *enterprise.UnpauseRequest) (*enterprise.UnpauseResponse, error)

type mockActivateEnterprise struct{ handler activateEnterpriseFunc }
type mockGetState struct{ handler getStateFunc }
type mockGetActivationCode struct{ handler getActivationCodeFunc }
type mockDeactivateEnterprise struct{ handler deactivateEnterpriseFunc }
type mockHeartbeatEnterprise struct{ handler heartbeatEnterpriseFunc }
type mockPause struct{ handler pauseFunc }
type mockPauseStatus struct{ handler pauseStatusFunc }
type mockUnpause struct{ handler unpauseFunc }

func (mock *mockActivateEnterprise) Use(cb activateEnterpriseFunc)     { mock.handler = cb }
func (mock *mockGetState) Use(cb getStateFunc)                         { mock.handler = cb }
func (mock *mockGetActivationCode) Use(cb getActivationCodeFunc)       { mock.handler = cb }
func (mock *mockDeactivateEnterprise) Use(cb deactivateEnterpriseFunc) { mock.handler = cb }
func (mock *mockHeartbeatEnterprise) Use(cb heartbeatEnterpriseFunc)   { mock.handler = cb }
func (mock *mockPause) Use(cb pauseFunc)                               { mock.handler = cb }
func (mock *mockPauseStatus) Use(cb pauseStatusFunc)                   { mock.handler = cb }
func (mock *mockUnpause) Use(cb unpauseFunc)                           { mock.handler = cb }

type enterpriseServerAPI struct {
	mock *mockEnterpriseServer
}

type mockEnterpriseServer struct {
	api               enterpriseServerAPI
	Activate          mockActivateEnterprise
	GetState          mockGetState
	GetActivationCode mockGetActivationCode
	Deactivate        mockDeactivateEnterprise
	Heartbeat         mockHeartbeatEnterprise
	Pause             mockPause
	PauseStatus       mockPauseStatus
	Unpause           mockUnpause
}

func (api *enterpriseServerAPI) Activate(ctx context.Context, req *enterprise.ActivateRequest) (*enterprise.ActivateResponse, error) {
	if api.mock.Activate.handler != nil {
		return api.mock.Activate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock enterprise.Activate")
}
func (api *enterpriseServerAPI) GetState(ctx context.Context, req *enterprise.GetStateRequest) (*enterprise.GetStateResponse, error) {
	if api.mock.GetState.handler != nil {
		return api.mock.GetState.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock enterprise.GetState")
}
func (api *enterpriseServerAPI) GetActivationCode(ctx context.Context, req *enterprise.GetActivationCodeRequest) (*enterprise.GetActivationCodeResponse, error) {
	if api.mock.GetActivationCode.handler != nil {
		return api.mock.GetActivationCode.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock enterprise.GetActivationCode")
}
func (api *enterpriseServerAPI) Deactivate(ctx context.Context, req *enterprise.DeactivateRequest) (*enterprise.DeactivateResponse, error) {
	if api.mock.Deactivate.handler != nil {
		return api.mock.Deactivate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock enterprise.Deactivate")
}
func (api *enterpriseServerAPI) Heartbeat(ctx context.Context, req *enterprise.HeartbeatRequest) (*enterprise.HeartbeatResponse, error) {
	if api.mock.Heartbeat.handler != nil {
		return api.mock.Heartbeat.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock enterprise.Heartbeat")
}
func (api *enterpriseServerAPI) Pause(ctx context.Context, req *enterprise.PauseRequest) (*enterprise.PauseResponse, error) {
	if api.mock.Heartbeat.handler != nil {
		return api.mock.Pause.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock enterprise.Pause")
}
func (api *enterpriseServerAPI) PauseStatus(ctx context.Context, req *enterprise.PauseStatusRequest) (*enterprise.PauseStatusResponse, error) {
	if api.mock.Heartbeat.handler != nil {
		return api.mock.PauseStatus.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock enterprise.PauseStatus")
}
func (api *enterpriseServerAPI) Unpause(ctx context.Context, req *enterprise.UnpauseRequest) (*enterprise.UnpauseResponse, error) {
	if api.mock.Heartbeat.handler != nil {
		return api.mock.Unpause.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock enterprise.Unpause")
}

/* PFS Server Mocks */

type activateAuthPFSFunc func(context.Context, *pfs.ActivateAuthRequest) (*pfs.ActivateAuthResponse, error)
type createRepoFunc func(context.Context, *pfs.CreateRepoRequest) (*types.Empty, error)
type inspectRepoFunc func(context.Context, *pfs.InspectRepoRequest) (*pfs.RepoInfo, error)
type listRepoFunc func(*pfs.ListRepoRequest, pfs.API_ListRepoServer) error
type deleteRepoFunc func(context.Context, *pfs.DeleteRepoRequest) (*types.Empty, error)
type startCommitFunc func(context.Context, *pfs.StartCommitRequest) (*pfs.Commit, error)
type finishCommitFunc func(context.Context, *pfs.FinishCommitRequest) (*types.Empty, error)
type inspectCommitFunc func(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error)
type listCommitFunc func(*pfs.ListCommitRequest, pfs.API_ListCommitServer) error
type squashCommitSetFunc func(context.Context, *pfs.SquashCommitSetRequest) (*types.Empty, error)
type dropCommitSetFunc func(context.Context, *pfs.DropCommitSetRequest) (*types.Empty, error)
type inspectCommitSetFunc func(*pfs.InspectCommitSetRequest, pfs.API_InspectCommitSetServer) error
type listCommitSetFunc func(*pfs.ListCommitSetRequest, pfs.API_ListCommitSetServer) error
type subscribeCommitFunc func(*pfs.SubscribeCommitRequest, pfs.API_SubscribeCommitServer) error
type clearCommitFunc func(context.Context, *pfs.ClearCommitRequest) (*types.Empty, error)
type createBranchFunc func(context.Context, *pfs.CreateBranchRequest) (*types.Empty, error)
type inspectBranchFunc func(context.Context, *pfs.InspectBranchRequest) (*pfs.BranchInfo, error)
type listBranchFunc func(*pfs.ListBranchRequest, pfs.API_ListBranchServer) error
type deleteBranchFunc func(context.Context, *pfs.DeleteBranchRequest) (*types.Empty, error)
type createProjectFunc func(context.Context, *pfs.CreateProjectRequest) (*types.Empty, error)
type inspectProjectFunc func(context.Context, *pfs.InspectProjectRequest) (*pfs.ProjectInfo, error)
type listProjectFunc func(*pfs.ListProjectRequest, pfs.API_ListProjectServer) error
type deleteProjectFunc func(context.Context, *pfs.DeleteProjectRequest) (*types.Empty, error)
type modifyFileFunc func(pfs.API_ModifyFileServer) error
type getFileTARFunc func(*pfs.GetFileRequest, pfs.API_GetFileTARServer) error
type getFileFunc func(*pfs.GetFileRequest, pfs.API_GetFileServer) error
type inspectFileFunc func(context.Context, *pfs.InspectFileRequest) (*pfs.FileInfo, error)
type listFileFunc func(*pfs.ListFileRequest, pfs.API_ListFileServer) error
type walkFileFunc func(*pfs.WalkFileRequest, pfs.API_WalkFileServer) error
type globFileFunc func(*pfs.GlobFileRequest, pfs.API_GlobFileServer) error
type diffFileFunc func(*pfs.DiffFileRequest, pfs.API_DiffFileServer) error
type deleteAllPFSFunc func(context.Context, *types.Empty) (*types.Empty, error)
type fsckFunc func(*pfs.FsckRequest, pfs.API_FsckServer) error
type createFileSetFunc func(pfs.API_CreateFileSetServer) error
type addFileSetFunc func(context.Context, *pfs.AddFileSetRequest) (*types.Empty, error)
type getFileSetFunc func(context.Context, *pfs.GetFileSetRequest) (*pfs.CreateFileSetResponse, error)
type renewFileSetFunc func(context.Context, *pfs.RenewFileSetRequest) (*types.Empty, error)
type composeFileSetFunc func(context.Context, *pfs.ComposeFileSetRequest) (*pfs.CreateFileSetResponse, error)
type shardFileSetFunc func(context.Context, *pfs.ShardFileSetRequest) (*pfs.ShardFileSetResponse, error)
type checkStorageFunc func(context.Context, *pfs.CheckStorageRequest) (*pfs.CheckStorageResponse, error)
type putCacheFunc func(context.Context, *pfs.PutCacheRequest) (*types.Empty, error)
type getCacheFunc func(context.Context, *pfs.GetCacheRequest) (*pfs.GetCacheResponse, error)
type clearCacheFunc func(context.Context, *pfs.ClearCacheRequest) (*types.Empty, error)
type runLoadTestFunc func(context.Context, *pfs.RunLoadTestRequest) (*pfs.RunLoadTestResponse, error)
type runLoadTestDefaultFunc func(context.Context, *types.Empty) (*pfs.RunLoadTestResponse, error)
type listTaskPFSFunc func(*task.ListTaskRequest, pfs.API_ListTaskServer) error
type egressFunc func(context.Context, *pfs.EgressRequest) (*pfs.EgressResponse, error)

type mockActivateAuthPFS struct{ handler activateAuthPFSFunc }
type mockCreateRepo struct{ handler createRepoFunc }
type mockInspectRepo struct{ handler inspectRepoFunc }
type mockListRepo struct{ handler listRepoFunc }
type mockDeleteRepo struct{ handler deleteRepoFunc }
type mockStartCommit struct{ handler startCommitFunc }
type mockFinishCommit struct{ handler finishCommitFunc }
type mockInspectCommit struct{ handler inspectCommitFunc }
type mockListCommit struct{ handler listCommitFunc }
type mockSquashCommitSet struct{ handler squashCommitSetFunc }
type mockDropCommitSet struct{ handler dropCommitSetFunc }
type mockInspectCommitSet struct{ handler inspectCommitSetFunc }
type mockListCommitSet struct{ handler listCommitSetFunc }
type mockSubscribeCommit struct{ handler subscribeCommitFunc }
type mockClearCommit struct{ handler clearCommitFunc }
type mockCreateBranch struct{ handler createBranchFunc }
type mockInspectBranch struct{ handler inspectBranchFunc }
type mockListBranch struct{ handler listBranchFunc }
type mockDeleteBranch struct{ handler deleteBranchFunc }
type mockCreateProject struct{ handler createProjectFunc }
type mockInspectProject struct{ handler inspectProjectFunc }
type mockListProject struct{ handler listProjectFunc }
type mockDeleteProject struct{ handler deleteProjectFunc }
type mockModifyFile struct{ handler modifyFileFunc }
type mockGetFile struct{ handler getFileFunc }
type mockGetFileTAR struct{ handler getFileTARFunc }
type mockInspectFile struct{ handler inspectFileFunc }
type mockListFile struct{ handler listFileFunc }
type mockWalkFile struct{ handler walkFileFunc }
type mockGlobFile struct{ handler globFileFunc }
type mockDiffFile struct{ handler diffFileFunc }
type mockDeleteAllPFS struct{ handler deleteAllPFSFunc }
type mockFsck struct{ handler fsckFunc }
type mockCreateFileSet struct{ handler createFileSetFunc }
type mockAddFileSet struct{ handler addFileSetFunc }
type mockGetFileSet struct{ handler getFileSetFunc }
type mockRenewFileSet struct{ handler renewFileSetFunc }
type mockComposeFileSet struct{ handler composeFileSetFunc }
type mockShardFileSet struct{ handler shardFileSetFunc }
type mockCheckStorage struct{ handler checkStorageFunc }
type mockPutCache struct{ handler putCacheFunc }
type mockGetCache struct{ handler getCacheFunc }
type mockClearCache struct{ handler clearCacheFunc }
type mockRunLoadTest struct{ handler runLoadTestFunc }
type mockRunLoadTestDefault struct{ handler runLoadTestDefaultFunc }
type mockListTaskPFS struct{ handler listTaskPFSFunc }
type mockEgress struct{ handler egressFunc }

func (mock *mockActivateAuthPFS) Use(cb activateAuthPFSFunc)       { mock.handler = cb }
func (mock *mockCreateRepo) Use(cb createRepoFunc)                 { mock.handler = cb }
func (mock *mockInspectRepo) Use(cb inspectRepoFunc)               { mock.handler = cb }
func (mock *mockListRepo) Use(cb listRepoFunc)                     { mock.handler = cb }
func (mock *mockDeleteRepo) Use(cb deleteRepoFunc)                 { mock.handler = cb }
func (mock *mockStartCommit) Use(cb startCommitFunc)               { mock.handler = cb }
func (mock *mockFinishCommit) Use(cb finishCommitFunc)             { mock.handler = cb }
func (mock *mockInspectCommit) Use(cb inspectCommitFunc)           { mock.handler = cb }
func (mock *mockListCommit) Use(cb listCommitFunc)                 { mock.handler = cb }
func (mock *mockSubscribeCommit) Use(cb subscribeCommitFunc)       { mock.handler = cb }
func (mock *mockClearCommit) Use(cb clearCommitFunc)               { mock.handler = cb }
func (mock *mockSquashCommitSet) Use(cb squashCommitSetFunc)       { mock.handler = cb }
func (mock *mockDropCommitSet) Use(cb dropCommitSetFunc)           { mock.handler = cb }
func (mock *mockInspectCommitSet) Use(cb inspectCommitSetFunc)     { mock.handler = cb }
func (mock *mockListCommitSet) Use(cb listCommitSetFunc)           { mock.handler = cb }
func (mock *mockCreateBranch) Use(cb createBranchFunc)             { mock.handler = cb }
func (mock *mockInspectBranch) Use(cb inspectBranchFunc)           { mock.handler = cb }
func (mock *mockListBranch) Use(cb listBranchFunc)                 { mock.handler = cb }
func (mock *mockDeleteBranch) Use(cb deleteBranchFunc)             { mock.handler = cb }
func (mock *mockCreateProject) Use(cb createProjectFunc)           { mock.handler = cb }
func (mock *mockInspectProject) Use(cb inspectProjectFunc)         { mock.handler = cb }
func (mock *mockListProject) Use(cb listProjectFunc)               { mock.handler = cb }
func (mock *mockDeleteProject) Use(cb deleteProjectFunc)           { mock.handler = cb }
func (mock *mockModifyFile) Use(cb modifyFileFunc)                 { mock.handler = cb }
func (mock *mockGetFile) Use(cb getFileFunc)                       { mock.handler = cb }
func (mock *mockGetFileTAR) Use(cb getFileTARFunc)                 { mock.handler = cb }
func (mock *mockInspectFile) Use(cb inspectFileFunc)               { mock.handler = cb }
func (mock *mockListFile) Use(cb listFileFunc)                     { mock.handler = cb }
func (mock *mockWalkFile) Use(cb walkFileFunc)                     { mock.handler = cb }
func (mock *mockGlobFile) Use(cb globFileFunc)                     { mock.handler = cb }
func (mock *mockDiffFile) Use(cb diffFileFunc)                     { mock.handler = cb }
func (mock *mockDeleteAllPFS) Use(cb deleteAllPFSFunc)             { mock.handler = cb }
func (mock *mockFsck) Use(cb fsckFunc)                             { mock.handler = cb }
func (mock *mockCreateFileSet) Use(cb createFileSetFunc)           { mock.handler = cb }
func (mock *mockAddFileSet) Use(cb addFileSetFunc)                 { mock.handler = cb }
func (mock *mockGetFileSet) Use(cb getFileSetFunc)                 { mock.handler = cb }
func (mock *mockRenewFileSet) Use(cb renewFileSetFunc)             { mock.handler = cb }
func (mock *mockComposeFileSet) Use(cb composeFileSetFunc)         { mock.handler = cb }
func (mock *mockShardFileSet) Use(cb shardFileSetFunc)             { mock.handler = cb }
func (mock *mockCheckStorage) Use(cb checkStorageFunc)             { mock.handler = cb }
func (mock *mockPutCache) Use(cb putCacheFunc)                     { mock.handler = cb }
func (mock *mockGetCache) Use(cb getCacheFunc)                     { mock.handler = cb }
func (mock *mockClearCache) Use(cb clearCacheFunc)                 { mock.handler = cb }
func (mock *mockRunLoadTest) Use(cb runLoadTestFunc)               { mock.handler = cb }
func (mock *mockRunLoadTestDefault) Use(cb runLoadTestDefaultFunc) { mock.handler = cb }
func (mock *mockListTaskPFS) Use(cb listTaskPFSFunc)               { mock.handler = cb }
func (mock *mockEgress) Use(cb egressFunc)                         { mock.handler = cb }

type pfsServerAPI struct {
	mock *mockPFSServer
}

type mockPFSServer struct {
	api                pfsServerAPI
	ActivateAuth       mockActivateAuthPFS
	CreateRepo         mockCreateRepo
	InspectRepo        mockInspectRepo
	ListRepo           mockListRepo
	DeleteRepo         mockDeleteRepo
	StartCommit        mockStartCommit
	FinishCommit       mockFinishCommit
	InspectCommit      mockInspectCommit
	ListCommit         mockListCommit
	SubscribeCommit    mockSubscribeCommit
	ClearCommit        mockClearCommit
	SquashCommitSet    mockSquashCommitSet
	DropCommitSet      mockDropCommitSet
	InspectCommitSet   mockInspectCommitSet
	ListCommitSet      mockListCommitSet
	CreateBranch       mockCreateBranch
	InspectBranch      mockInspectBranch
	ListBranch         mockListBranch
	DeleteBranch       mockDeleteBranch
	CreateProject      mockCreateProject
	InspectProject     mockInspectProject
	ListProject        mockListProject
	DeleteProject      mockDeleteProject
	ModifyFile         mockModifyFile
	GetFile            mockGetFile
	GetFileTAR         mockGetFileTAR
	InspectFile        mockInspectFile
	ListFile           mockListFile
	WalkFile           mockWalkFile
	GlobFile           mockGlobFile
	DiffFile           mockDiffFile
	DeleteAll          mockDeleteAllPFS
	Fsck               mockFsck
	CreateFileSet      mockCreateFileSet
	AddFileSet         mockAddFileSet
	GetFileSet         mockGetFileSet
	RenewFileSet       mockRenewFileSet
	ComposeFileSet     mockComposeFileSet
	ShardFileSet       mockShardFileSet
	CheckStorage       mockCheckStorage
	PutCache           mockPutCache
	GetCache           mockGetCache
	ClearCache         mockClearCache
	RunLoadTest        mockRunLoadTest
	RunLoadTestDefault mockRunLoadTestDefault
	ListTask           mockListTaskPFS
	Egress             mockEgress
}

func (api *pfsServerAPI) ActivateAuth(ctx context.Context, req *pfs.ActivateAuthRequest) (*pfs.ActivateAuthResponse, error) {
	if api.mock.ActivateAuth.handler != nil {
		return api.mock.ActivateAuth.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.ActivateAuth")
}
func (api *pfsServerAPI) CreateRepo(ctx context.Context, req *pfs.CreateRepoRequest) (*types.Empty, error) {
	if api.mock.CreateRepo.handler != nil {
		return api.mock.CreateRepo.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.CreateRepo")
}
func (api *pfsServerAPI) InspectRepo(ctx context.Context, req *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	if api.mock.InspectRepo.handler != nil {
		return api.mock.InspectRepo.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.InspectRepo")
}
func (api *pfsServerAPI) ListRepo(req *pfs.ListRepoRequest, srv pfs.API_ListRepoServer) error {
	if api.mock.ListRepo.handler != nil {
		return api.mock.ListRepo.handler(req, srv)
	}
	return errors.Errorf("unhandled pachd mock pfs.ListRepo")
}
func (api *pfsServerAPI) DeleteRepo(ctx context.Context, req *pfs.DeleteRepoRequest) (*types.Empty, error) {
	if api.mock.DeleteRepo.handler != nil {
		return api.mock.DeleteRepo.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.DeleteRepo")
}
func (api *pfsServerAPI) StartCommit(ctx context.Context, req *pfs.StartCommitRequest) (*pfs.Commit, error) {
	if api.mock.StartCommit.handler != nil {
		return api.mock.StartCommit.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.StartCommit")
}
func (api *pfsServerAPI) FinishCommit(ctx context.Context, req *pfs.FinishCommitRequest) (*types.Empty, error) {
	if api.mock.FinishCommit.handler != nil {
		return api.mock.FinishCommit.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.FinishCommit")
}
func (api *pfsServerAPI) InspectCommit(ctx context.Context, req *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	if api.mock.InspectCommit.handler != nil {
		return api.mock.InspectCommit.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.InspectCommit")
}
func (api *pfsServerAPI) ListCommit(req *pfs.ListCommitRequest, serv pfs.API_ListCommitServer) error {
	if api.mock.ListCommit.handler != nil {
		return api.mock.ListCommit.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.ListCommit")
}
func (api *pfsServerAPI) SquashCommitSet(ctx context.Context, req *pfs.SquashCommitSetRequest) (*types.Empty, error) {
	if api.mock.SquashCommitSet.handler != nil {
		return api.mock.SquashCommitSet.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.SquashCommitSet")
}
func (api *pfsServerAPI) DropCommitSet(ctx context.Context, req *pfs.DropCommitSetRequest) (*types.Empty, error) {
	if api.mock.DropCommitSet.handler != nil {
		return api.mock.DropCommitSet.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.DropCommitSet")
}
func (api *pfsServerAPI) InspectCommitSet(req *pfs.InspectCommitSetRequest, serv pfs.API_InspectCommitSetServer) error {
	if api.mock.InspectCommitSet.handler != nil {
		return api.mock.InspectCommitSet.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.InspectCommitSet")
}
func (api *pfsServerAPI) ListCommitSet(req *pfs.ListCommitSetRequest, serv pfs.API_ListCommitSetServer) error {
	if api.mock.ListCommitSet.handler != nil {
		return api.mock.ListCommitSet.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.ListCommitSet")
}
func (api *pfsServerAPI) SubscribeCommit(req *pfs.SubscribeCommitRequest, serv pfs.API_SubscribeCommitServer) error {
	if api.mock.SubscribeCommit.handler != nil {
		return api.mock.SubscribeCommit.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.SubscribeCommit")
}
func (api *pfsServerAPI) ClearCommit(ctx context.Context, req *pfs.ClearCommitRequest) (*types.Empty, error) {
	if api.mock.ClearCommit.handler != nil {
		return api.mock.ClearCommit.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.ClearCommit")
}
func (api *pfsServerAPI) CreateBranch(ctx context.Context, req *pfs.CreateBranchRequest) (*types.Empty, error) {
	if api.mock.CreateBranch.handler != nil {
		return api.mock.CreateBranch.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.CreateBranch")
}
func (api *pfsServerAPI) InspectBranch(ctx context.Context, req *pfs.InspectBranchRequest) (*pfs.BranchInfo, error) {
	if api.mock.InspectBranch.handler != nil {
		return api.mock.InspectBranch.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.InspectBranch")
}
func (api *pfsServerAPI) ListBranch(req *pfs.ListBranchRequest, srv pfs.API_ListBranchServer) error {
	if api.mock.ListBranch.handler != nil {
		return api.mock.ListBranch.handler(req, srv)
	}
	return errors.Errorf("unhandled pachd mock pfs.ListBranch")
}
func (api *pfsServerAPI) DeleteBranch(ctx context.Context, req *pfs.DeleteBranchRequest) (*types.Empty, error) {
	if api.mock.DeleteBranch.handler != nil {
		return api.mock.DeleteBranch.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.DeleteBranch")
}
func (api *pfsServerAPI) CreateProject(ctx context.Context, req *pfs.CreateProjectRequest) (*types.Empty, error) {
	if api.mock.CreateProject.handler != nil {
		return api.mock.CreateProject.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.CreateProject")
}
func (api *pfsServerAPI) InspectProject(ctx context.Context, req *pfs.InspectProjectRequest) (*pfs.ProjectInfo, error) {
	if api.mock.InspectProject.handler != nil {
		return api.mock.InspectProject.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.InspectProject")
}
func (api *pfsServerAPI) ListProject(req *pfs.ListProjectRequest, srv pfs.API_ListProjectServer) error {
	if api.mock.ListProject.handler != nil {
		return api.mock.ListProject.handler(req, srv)
	}
	return errors.Errorf("unhandled pachd mock pfs.ListProject")
}
func (api *pfsServerAPI) DeleteProject(ctx context.Context, req *pfs.DeleteProjectRequest) (*types.Empty, error) {
	if api.mock.DeleteProject.handler != nil {
		return api.mock.DeleteProject.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.DeleteProject")
}
func (api *pfsServerAPI) ModifyFile(serv pfs.API_ModifyFileServer) error {
	if api.mock.ModifyFile.handler != nil {
		return api.mock.ModifyFile.handler(serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.ModifyFile")
}
func (api *pfsServerAPI) GetFile(req *pfs.GetFileRequest, serv pfs.API_GetFileServer) error {
	if api.mock.GetFile.handler != nil {
		return api.mock.GetFile.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.GetFile")
}
func (api *pfsServerAPI) GetFileTAR(req *pfs.GetFileRequest, serv pfs.API_GetFileTARServer) error {
	if api.mock.GetFileTAR.handler != nil {
		return api.mock.GetFileTAR.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.GetFileTAR")
}
func (api *pfsServerAPI) InspectFile(ctx context.Context, req *pfs.InspectFileRequest) (*pfs.FileInfo, error) {
	if api.mock.InspectFile.handler != nil {
		return api.mock.InspectFile.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.InspectFile")
}
func (api *pfsServerAPI) ListFile(req *pfs.ListFileRequest, serv pfs.API_ListFileServer) error {
	if api.mock.ListFile.handler != nil {
		return api.mock.ListFile.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.ListFile")
}
func (api *pfsServerAPI) WalkFile(req *pfs.WalkFileRequest, serv pfs.API_WalkFileServer) error {
	if api.mock.WalkFile.handler != nil {
		return api.mock.WalkFile.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.WalkFile")
}
func (api *pfsServerAPI) GlobFile(req *pfs.GlobFileRequest, serv pfs.API_GlobFileServer) error {
	if api.mock.GlobFile.handler != nil {
		return api.mock.GlobFile.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.GlobFile")
}
func (api *pfsServerAPI) DiffFile(req *pfs.DiffFileRequest, serv pfs.API_DiffFileServer) error {
	if api.mock.DiffFile.handler != nil {
		return api.mock.DiffFile.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.DiffFile")
}
func (api *pfsServerAPI) DeleteAll(ctx context.Context, req *types.Empty) (*types.Empty, error) {
	if api.mock.DeleteAll.handler != nil {
		return api.mock.DeleteAll.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.DeleteAll")
}
func (api *pfsServerAPI) Fsck(req *pfs.FsckRequest, serv pfs.API_FsckServer) error {
	if api.mock.Fsck.handler != nil {
		return api.mock.Fsck.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.Fsck")
}
func (api *pfsServerAPI) CreateFileSet(srv pfs.API_CreateFileSetServer) error {
	if api.mock.CreateFileSet.handler != nil {
		return api.mock.CreateFileSet.handler(srv)
	}
	return errors.Errorf("unhandled pachd mock pfs.CreateFileSet")
}
func (api *pfsServerAPI) AddFileSet(ctx context.Context, req *pfs.AddFileSetRequest) (*types.Empty, error) {
	if api.mock.AddFileSet.handler != nil {
		return api.mock.AddFileSet.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.AddFileSet")
}
func (api *pfsServerAPI) GetFileSet(ctx context.Context, req *pfs.GetFileSetRequest) (*pfs.CreateFileSetResponse, error) {
	if api.mock.AddFileSet.handler != nil {
		return api.mock.GetFileSet.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.AddFileSet")
}
func (api *pfsServerAPI) RenewFileSet(ctx context.Context, req *pfs.RenewFileSetRequest) (*types.Empty, error) {
	if api.mock.RenewFileSet.handler != nil {
		return api.mock.RenewFileSet.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.RenewFileSet")
}
func (api *pfsServerAPI) ComposeFileSet(ctx context.Context, req *pfs.ComposeFileSetRequest) (*pfs.CreateFileSetResponse, error) {
	if api.mock.ComposeFileSet.handler != nil {
		return api.mock.ComposeFileSet.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.ComposeFileSet")
}
func (api *pfsServerAPI) ShardFileSet(ctx context.Context, req *pfs.ShardFileSetRequest) (*pfs.ShardFileSetResponse, error) {
	if api.mock.ShardFileSet.handler != nil {
		return api.mock.ShardFileSet.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.ShardFileSet")
}
func (api *pfsServerAPI) CheckStorage(ctx context.Context, req *pfs.CheckStorageRequest) (*pfs.CheckStorageResponse, error) {
	if api.mock.CheckStorage.handler != nil {
		return api.mock.CheckStorage.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock CheckStorage")
}
func (api *pfsServerAPI) PutCache(ctx context.Context, req *pfs.PutCacheRequest) (*types.Empty, error) {
	if api.mock.PutCache.handler != nil {
		return api.mock.PutCache.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock PutCache")
}
func (api *pfsServerAPI) GetCache(ctx context.Context, req *pfs.GetCacheRequest) (*pfs.GetCacheResponse, error) {
	if api.mock.GetCache.handler != nil {
		return api.mock.GetCache.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock GetCache")
}
func (api *pfsServerAPI) ClearCache(ctx context.Context, req *pfs.ClearCacheRequest) (*types.Empty, error) {
	if api.mock.ClearCache.handler != nil {
		return api.mock.ClearCache.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock ClearCache")
}
func (api *pfsServerAPI) RunLoadTest(ctx context.Context, req *pfs.RunLoadTestRequest) (*pfs.RunLoadTestResponse, error) {
	if api.mock.RunLoadTest.handler != nil {
		return api.mock.RunLoadTest.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.RunLoadTest")
}
func (api *pfsServerAPI) RunLoadTestDefault(ctx context.Context, req *types.Empty) (*pfs.RunLoadTestResponse, error) {
	if api.mock.RunLoadTestDefault.handler != nil {
		return api.mock.RunLoadTestDefault.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.RunLoadTestDefault")
}
func (api *pfsServerAPI) ListTask(req *task.ListTaskRequest, server pfs.API_ListTaskServer) error {
	if api.mock.ListTask.handler != nil {
		return api.mock.ListTask.handler(req, server)
	}
	return errors.Errorf("unhandled pachd mock pfs.ListTask")
}
func (api *pfsServerAPI) Egress(ctx context.Context, req *pfs.EgressRequest) (*pfs.EgressResponse, error) {
	if api.mock.Egress.handler != nil {
		return api.mock.Egress.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.Egress")
}

/* PPS Server Mocks */

type inspectJobFunc func(context.Context, *pps.InspectJobRequest) (*pps.JobInfo, error)
type listJobFunc func(*pps.ListJobRequest, pps.API_ListJobServer) error
type subscribeJobFunc func(*pps.SubscribeJobRequest, pps.API_SubscribeJobServer) error
type deleteJobFunc func(context.Context, *pps.DeleteJobRequest) (*types.Empty, error)
type stopJobFunc func(context.Context, *pps.StopJobRequest) (*types.Empty, error)
type updateJobStateFunc func(context.Context, *pps.UpdateJobStateRequest) (*types.Empty, error)
type inspectJobSetFunc func(*pps.InspectJobSetRequest, pps.API_InspectJobSetServer) error
type listJobSetFunc func(*pps.ListJobSetRequest, pps.API_ListJobSetServer) error
type inspectDatumFunc func(context.Context, *pps.InspectDatumRequest) (*pps.DatumInfo, error)
type listDatumFunc func(*pps.ListDatumRequest, pps.API_ListDatumServer) error
type restartDatumFunc func(context.Context, *pps.RestartDatumRequest) (*types.Empty, error)
type createPipelineFunc func(context.Context, *pps.CreatePipelineRequest) (*types.Empty, error)
type inspectPipelineFunc func(context.Context, *pps.InspectPipelineRequest) (*pps.PipelineInfo, error)
type listPipelineFunc func(*pps.ListPipelineRequest, pps.API_ListPipelineServer) error
type deletePipelineFunc func(context.Context, *pps.DeletePipelineRequest) (*types.Empty, error)
type startPipelineFunc func(context.Context, *pps.StartPipelineRequest) (*types.Empty, error)
type stopPipelineFunc func(context.Context, *pps.StopPipelineRequest) (*types.Empty, error)
type runPipelineFunc func(context.Context, *pps.RunPipelineRequest) (*types.Empty, error)
type runCronFunc func(context.Context, *pps.RunCronRequest) (*types.Empty, error)
type createSecretFunc func(context.Context, *pps.CreateSecretRequest) (*types.Empty, error)
type deleteSecretFunc func(context.Context, *pps.DeleteSecretRequest) (*types.Empty, error)
type inspectSecretFunc func(context.Context, *pps.InspectSecretRequest) (*pps.SecretInfo, error)
type listSecretFunc func(context.Context, *types.Empty) (*pps.SecretInfos, error)
type deleteAllPPSFunc func(context.Context, *types.Empty) (*types.Empty, error)
type getLogsFunc func(*pps.GetLogsRequest, pps.API_GetLogsServer) error
type activateAuthPPSFunc func(context.Context, *pps.ActivateAuthRequest) (*pps.ActivateAuthResponse, error)
type runLoadTestPPSFunc func(context.Context, *pps.RunLoadTestRequest) (*pps.RunLoadTestResponse, error)
type runLoadTestDefaultPPSFunc func(context.Context, *types.Empty) (*pps.RunLoadTestResponse, error)
type renderTemplateFunc func(context.Context, *pps.RenderTemplateRequest) (*pps.RenderTemplateResponse, error)
type listTaskPPSFunc func(*task.ListTaskRequest, pps.API_ListTaskServer) error

type mockInspectJob struct{ handler inspectJobFunc }
type mockListJob struct{ handler listJobFunc }
type mockSubscribeJob struct{ handler subscribeJobFunc }
type mockDeleteJob struct{ handler deleteJobFunc }
type mockStopJob struct{ handler stopJobFunc }
type mockUpdateJobState struct{ handler updateJobStateFunc }
type mockInspectJobSet struct{ handler inspectJobSetFunc }
type mockListJobSet struct{ handler listJobSetFunc }
type mockInspectDatum struct{ handler inspectDatumFunc }
type mockListDatum struct{ handler listDatumFunc }
type mockRestartDatum struct{ handler restartDatumFunc }
type mockCreatePipeline struct{ handler createPipelineFunc }
type mockInspectPipeline struct{ handler inspectPipelineFunc }
type mockListPipeline struct{ handler listPipelineFunc }
type mockDeletePipeline struct{ handler deletePipelineFunc }
type mockStartPipeline struct{ handler startPipelineFunc }
type mockStopPipeline struct{ handler stopPipelineFunc }
type mockRunPipeline struct{ handler runPipelineFunc }
type mockRunCron struct{ handler runCronFunc }
type mockCreateSecret struct{ handler createSecretFunc }
type mockDeleteSecret struct{ handler deleteSecretFunc }
type mockInspectSecret struct{ handler inspectSecretFunc }
type mockListSecret struct{ handler listSecretFunc }
type mockDeleteAllPPS struct{ handler deleteAllPPSFunc }
type mockGetLogs struct{ handler getLogsFunc }
type mockActivateAuthPPS struct{ handler activateAuthPPSFunc }
type mockRunLoadTestPPS struct{ handler runLoadTestPPSFunc }
type mockRunLoadTestDefaultPPS struct{ handler runLoadTestDefaultPPSFunc }
type mockRenderTemplate struct{ handler renderTemplateFunc }
type mockListTaskPPS struct{ handler listTaskPPSFunc }

func (mock *mockInspectJob) Use(cb inspectJobFunc)                       { mock.handler = cb }
func (mock *mockListJob) Use(cb listJobFunc)                             { mock.handler = cb }
func (mock *mockSubscribeJob) Use(cb subscribeJobFunc)                   { mock.handler = cb }
func (mock *mockDeleteJob) Use(cb deleteJobFunc)                         { mock.handler = cb }
func (mock *mockStopJob) Use(cb stopJobFunc)                             { mock.handler = cb }
func (mock *mockUpdateJobState) Use(cb updateJobStateFunc)               { mock.handler = cb }
func (mock *mockInspectJobSet) Use(cb inspectJobSetFunc)                 { mock.handler = cb }
func (mock *mockListJobSet) Use(cb listJobSetFunc)                       { mock.handler = cb }
func (mock *mockInspectDatum) Use(cb inspectDatumFunc)                   { mock.handler = cb }
func (mock *mockListDatum) Use(cb listDatumFunc)                         { mock.handler = cb }
func (mock *mockRestartDatum) Use(cb restartDatumFunc)                   { mock.handler = cb }
func (mock *mockCreatePipeline) Use(cb createPipelineFunc)               { mock.handler = cb }
func (mock *mockInspectPipeline) Use(cb inspectPipelineFunc)             { mock.handler = cb }
func (mock *mockListPipeline) Use(cb listPipelineFunc)                   { mock.handler = cb }
func (mock *mockDeletePipeline) Use(cb deletePipelineFunc)               { mock.handler = cb }
func (mock *mockStartPipeline) Use(cb startPipelineFunc)                 { mock.handler = cb }
func (mock *mockStopPipeline) Use(cb stopPipelineFunc)                   { mock.handler = cb }
func (mock *mockRunPipeline) Use(cb runPipelineFunc)                     { mock.handler = cb }
func (mock *mockRunCron) Use(cb runCronFunc)                             { mock.handler = cb }
func (mock *mockCreateSecret) Use(cb createSecretFunc)                   { mock.handler = cb }
func (mock *mockDeleteSecret) Use(cb deleteSecretFunc)                   { mock.handler = cb }
func (mock *mockInspectSecret) Use(cb inspectSecretFunc)                 { mock.handler = cb }
func (mock *mockListSecret) Use(cb listSecretFunc)                       { mock.handler = cb }
func (mock *mockDeleteAllPPS) Use(cb deleteAllPPSFunc)                   { mock.handler = cb }
func (mock *mockGetLogs) Use(cb getLogsFunc)                             { mock.handler = cb }
func (mock *mockActivateAuthPPS) Use(cb activateAuthPPSFunc)             { mock.handler = cb }
func (mock *mockRunLoadTestPPS) Use(cb runLoadTestPPSFunc)               { mock.handler = cb }
func (mock *mockRunLoadTestDefaultPPS) Use(cb runLoadTestDefaultPPSFunc) { mock.handler = cb }
func (mock *mockRenderTemplate) Use(cb renderTemplateFunc)               { mock.handler = cb }
func (mock *mockListTaskPPS) Use(cb listTaskPPSFunc)                     { mock.handler = cb }

type ppsServerAPI struct {
	mock *mockPPSServer
}

type mockPPSServer struct {
	api                ppsServerAPI
	InspectJob         mockInspectJob
	ListJob            mockListJob
	SubscribeJob       mockSubscribeJob
	DeleteJob          mockDeleteJob
	StopJob            mockStopJob
	UpdateJobState     mockUpdateJobState
	InspectJobSet      mockInspectJobSet
	ListJobSet         mockListJobSet
	InspectDatum       mockInspectDatum
	ListDatum          mockListDatum
	RestartDatum       mockRestartDatum
	CreatePipeline     mockCreatePipeline
	InspectPipeline    mockInspectPipeline
	ListPipeline       mockListPipeline
	DeletePipeline     mockDeletePipeline
	StartPipeline      mockStartPipeline
	StopPipeline       mockStopPipeline
	RunPipeline        mockRunPipeline
	RunCron            mockRunCron
	CreateSecret       mockCreateSecret
	DeleteSecret       mockDeleteSecret
	InspectSecret      mockInspectSecret
	ListSecret         mockListSecret
	DeleteAll          mockDeleteAllPPS
	GetLogs            mockGetLogs
	ActivateAuth       mockActivateAuthPPS
	RunLoadTest        mockRunLoadTestPPS
	RunLoadTestDefault mockRunLoadTestDefaultPPS
	RenderTemplate     mockRenderTemplate
	ListTask           mockListTaskPPS
}

func (api *ppsServerAPI) InspectJob(ctx context.Context, req *pps.InspectJobRequest) (*pps.JobInfo, error) {
	if api.mock.InspectJob.handler != nil {
		return api.mock.InspectJob.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.InspectJob")
}
func (api *ppsServerAPI) ListJob(req *pps.ListJobRequest, serv pps.API_ListJobServer) error {
	if api.mock.ListJob.handler != nil {
		return api.mock.ListJob.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pps.ListJob")
}
func (api *ppsServerAPI) SubscribeJob(req *pps.SubscribeJobRequest, serv pps.API_SubscribeJobServer) error {
	if api.mock.SubscribeJob.handler != nil {
		return api.mock.SubscribeJob.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pps.SubscribeJob")
}
func (api *ppsServerAPI) DeleteJob(ctx context.Context, req *pps.DeleteJobRequest) (*types.Empty, error) {
	if api.mock.DeleteJob.handler != nil {
		return api.mock.DeleteJob.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.DeleteJob")
}
func (api *ppsServerAPI) UpdateJobState(ctx context.Context, req *pps.UpdateJobStateRequest) (*types.Empty, error) {
	if api.mock.UpdateJobState.handler != nil {
		return api.mock.UpdateJobState.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.UpdateJobState")
}
func (api *ppsServerAPI) StopJob(ctx context.Context, req *pps.StopJobRequest) (*types.Empty, error) {
	if api.mock.StopJob.handler != nil {
		return api.mock.StopJob.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.StopJob")
}
func (api *ppsServerAPI) InspectJobSet(req *pps.InspectJobSetRequest, serv pps.API_InspectJobSetServer) error {
	if api.mock.InspectJobSet.handler != nil {
		return api.mock.InspectJobSet.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pps.InspectJobSet")
}
func (api *ppsServerAPI) ListJobSet(req *pps.ListJobSetRequest, serv pps.API_ListJobSetServer) error {
	if api.mock.ListJobSet.handler != nil {
		return api.mock.ListJobSet.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pps.ListJobSet")
}
func (api *ppsServerAPI) InspectDatum(ctx context.Context, req *pps.InspectDatumRequest) (*pps.DatumInfo, error) {
	if api.mock.InspectDatum.handler != nil {
		return api.mock.InspectDatum.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.InspectDatum")
}
func (api *ppsServerAPI) ListDatum(req *pps.ListDatumRequest, serv pps.API_ListDatumServer) error {
	if api.mock.ListDatum.handler != nil {
		return api.mock.ListDatum.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pps.ListDatum")
}
func (api *ppsServerAPI) RestartDatum(ctx context.Context, req *pps.RestartDatumRequest) (*types.Empty, error) {
	if api.mock.RestartDatum.handler != nil {
		return api.mock.RestartDatum.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.RestartDatum")
}
func (api *ppsServerAPI) CreatePipeline(ctx context.Context, req *pps.CreatePipelineRequest) (*types.Empty, error) {
	if api.mock.CreatePipeline.handler != nil {
		return api.mock.CreatePipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.CreatePipeline")
}
func (api *ppsServerAPI) InspectPipeline(ctx context.Context, req *pps.InspectPipelineRequest) (*pps.PipelineInfo, error) {
	if api.mock.InspectPipeline.handler != nil {
		return api.mock.InspectPipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.InspectPipeline")
}
func (api *ppsServerAPI) ListPipeline(req *pps.ListPipelineRequest, srv pps.API_ListPipelineServer) error {
	if api.mock.ListPipeline.handler != nil {
		return api.mock.ListPipeline.handler(req, srv)
	}
	return errors.Errorf("unhandled pachd mock pps.ListPipeline")
}
func (api *ppsServerAPI) DeletePipeline(ctx context.Context, req *pps.DeletePipelineRequest) (*types.Empty, error) {
	if api.mock.DeletePipeline.handler != nil {
		return api.mock.DeletePipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.DeletePipeline")
}
func (api *ppsServerAPI) StartPipeline(ctx context.Context, req *pps.StartPipelineRequest) (*types.Empty, error) {
	if api.mock.StartPipeline.handler != nil {
		return api.mock.StartPipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.StartPipeline")
}
func (api *ppsServerAPI) StopPipeline(ctx context.Context, req *pps.StopPipelineRequest) (*types.Empty, error) {
	if api.mock.StopPipeline.handler != nil {
		return api.mock.StopPipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.StopPipeline")
}
func (api *ppsServerAPI) RunPipeline(ctx context.Context, req *pps.RunPipelineRequest) (*types.Empty, error) {
	if api.mock.RunPipeline.handler != nil {
		return api.mock.RunPipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.RunPipeline")
}
func (api *ppsServerAPI) RunCron(ctx context.Context, req *pps.RunCronRequest) (*types.Empty, error) {
	if api.mock.RunCron.handler != nil {
		return api.mock.RunCron.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.RunCron")
}
func (api *ppsServerAPI) CreateSecret(ctx context.Context, req *pps.CreateSecretRequest) (*types.Empty, error) {
	if api.mock.CreateSecret.handler != nil {
		return api.mock.CreateSecret.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.CreateSecret")
}
func (api *ppsServerAPI) DeleteSecret(ctx context.Context, req *pps.DeleteSecretRequest) (*types.Empty, error) {
	if api.mock.DeleteSecret.handler != nil {
		return api.mock.DeleteSecret.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.DeleteSecret")
}
func (api *ppsServerAPI) InspectSecret(ctx context.Context, req *pps.InspectSecretRequest) (*pps.SecretInfo, error) {
	if api.mock.InspectSecret.handler != nil {
		return api.mock.InspectSecret.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.InspectSecret")
}
func (api *ppsServerAPI) ListSecret(ctx context.Context, in *types.Empty) (*pps.SecretInfos, error) {
	if api.mock.ListSecret.handler != nil {
		return api.mock.ListSecret.handler(ctx, in)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.ListSecret")
}
func (api *ppsServerAPI) DeleteAll(ctx context.Context, req *types.Empty) (*types.Empty, error) {
	if api.mock.DeleteAll.handler != nil {
		return api.mock.DeleteAll.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.DeleteAll")
}
func (api *ppsServerAPI) GetLogs(req *pps.GetLogsRequest, serv pps.API_GetLogsServer) error {
	if api.mock.GetLogs.handler != nil {
		return api.mock.GetLogs.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pps.GetLogs")
}
func (api *ppsServerAPI) ActivateAuth(ctx context.Context, req *pps.ActivateAuthRequest) (*pps.ActivateAuthResponse, error) {
	if api.mock.ActivateAuth.handler != nil {
		return api.mock.ActivateAuth.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.ActivateAuth")
}
func (api *ppsServerAPI) RunLoadTest(ctx context.Context, req *pps.RunLoadTestRequest) (*pps.RunLoadTestResponse, error) {
	if api.mock.RunLoadTest.handler != nil {
		return api.mock.RunLoadTest.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.RunLoadTest")
}
func (api *ppsServerAPI) RunLoadTestDefault(ctx context.Context, req *types.Empty) (*pps.RunLoadTestResponse, error) {
	if api.mock.RunLoadTestDefault.handler != nil {
		return api.mock.RunLoadTestDefault.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.RunLoadTestDefault")
}
func (api *ppsServerAPI) RenderTemplate(ctx context.Context, req *pps.RenderTemplateRequest) (*pps.RenderTemplateResponse, error) {
	if api.mock.RenderTemplate.handler != nil {
		return api.mock.RenderTemplate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.RenderTemplate")
}
func (api *ppsServerAPI) ListTask(req *task.ListTaskRequest, server pps.API_ListTaskServer) error {
	if api.mock.ListTask.handler != nil {
		return api.mock.ListTask.handler(req, server)
	}
	return errors.Errorf("unhandled pachd mock pps.ListTask")
}

/* Transaction Server Mocks */

type batchTransactionFunc func(context.Context, *transaction.BatchTransactionRequest) (*transaction.TransactionInfo, error)
type startTransactionFunc func(context.Context, *transaction.StartTransactionRequest) (*transaction.Transaction, error)
type inspectTransactionFunc func(context.Context, *transaction.InspectTransactionRequest) (*transaction.TransactionInfo, error)
type deleteTransactionFunc func(context.Context, *transaction.DeleteTransactionRequest) (*types.Empty, error)
type listTransactionFunc func(context.Context, *transaction.ListTransactionRequest) (*transaction.TransactionInfos, error)
type finishTransactionFunc func(context.Context, *transaction.FinishTransactionRequest) (*transaction.TransactionInfo, error)
type deleteAllTransactionFunc func(context.Context, *transaction.DeleteAllRequest) (*types.Empty, error)

type mockBatchTransaction struct{ handler batchTransactionFunc }
type mockStartTransaction struct{ handler startTransactionFunc }
type mockInspectTransaction struct{ handler inspectTransactionFunc }
type mockDeleteTransaction struct{ handler deleteTransactionFunc }
type mockListTransaction struct{ handler listTransactionFunc }
type mockFinishTransaction struct{ handler finishTransactionFunc }
type mockDeleteAllTransaction struct{ handler deleteAllTransactionFunc }

func (mock *mockBatchTransaction) Use(cb batchTransactionFunc)         { mock.handler = cb }
func (mock *mockStartTransaction) Use(cb startTransactionFunc)         { mock.handler = cb }
func (mock *mockInspectTransaction) Use(cb inspectTransactionFunc)     { mock.handler = cb }
func (mock *mockDeleteTransaction) Use(cb deleteTransactionFunc)       { mock.handler = cb }
func (mock *mockListTransaction) Use(cb listTransactionFunc)           { mock.handler = cb }
func (mock *mockFinishTransaction) Use(cb finishTransactionFunc)       { mock.handler = cb }
func (mock *mockDeleteAllTransaction) Use(cb deleteAllTransactionFunc) { mock.handler = cb }

type transactionServerAPI struct {
	mock *mockTransactionServer
}

type mockTransactionServer struct {
	api                transactionServerAPI
	BatchTransaction   mockBatchTransaction
	StartTransaction   mockStartTransaction
	InspectTransaction mockInspectTransaction
	DeleteTransaction  mockDeleteTransaction
	ListTransaction    mockListTransaction
	FinishTransaction  mockFinishTransaction
	DeleteAll          mockDeleteAllTransaction
}

func (api *transactionServerAPI) BatchTransaction(ctx context.Context, req *transaction.BatchTransactionRequest) (*transaction.TransactionInfo, error) {
	if api.mock.BatchTransaction.handler != nil {
		return api.mock.BatchTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.BatchTransaction")
}
func (api *transactionServerAPI) StartTransaction(ctx context.Context, req *transaction.StartTransactionRequest) (*transaction.Transaction, error) {
	if api.mock.StartTransaction.handler != nil {
		return api.mock.StartTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.StartTransaction")
}
func (api *transactionServerAPI) InspectTransaction(ctx context.Context, req *transaction.InspectTransactionRequest) (*transaction.TransactionInfo, error) {
	if api.mock.InspectTransaction.handler != nil {
		return api.mock.InspectTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.InspectTransaction")
}
func (api *transactionServerAPI) DeleteTransaction(ctx context.Context, req *transaction.DeleteTransactionRequest) (*types.Empty, error) {
	if api.mock.DeleteTransaction.handler != nil {
		return api.mock.DeleteTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.DeleteTransaction")
}
func (api *transactionServerAPI) ListTransaction(ctx context.Context, req *transaction.ListTransactionRequest) (*transaction.TransactionInfos, error) {
	if api.mock.ListTransaction.handler != nil {
		return api.mock.ListTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.ListTransaction")
}
func (api *transactionServerAPI) FinishTransaction(ctx context.Context, req *transaction.FinishTransactionRequest) (*transaction.TransactionInfo, error) {
	if api.mock.FinishTransaction.handler != nil {
		return api.mock.FinishTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.FinishTransaction")
}
func (api *transactionServerAPI) DeleteAll(ctx context.Context, req *transaction.DeleteAllRequest) (*types.Empty, error) {
	if api.mock.DeleteAll.handler != nil {
		return api.mock.DeleteAll.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.DeleteAll")
}

/* Version Server Mocks */

type getVersionFunc func(context.Context, *types.Empty) (*version.Version, error)

type mockGetVersion struct{ handler getVersionFunc }

func (mock *mockGetVersion) Use(cb getVersionFunc) { mock.handler = cb }

type versionServerAPI struct {
	mock *mockVersionServer
}

type mockVersionServer struct {
	api        versionServerAPI
	GetVersion mockGetVersion
}

func (api *versionServerAPI) GetVersion(ctx context.Context, req *types.Empty) (*version.Version, error) {
	if api.mock.GetVersion.handler != nil {
		return api.mock.GetVersion.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock version.GetVersion")
}

/* Proxy Server Mocks */

type listenFunc func(*proxy.ListenRequest, proxy.API_ListenServer) error

type mockListen struct{ handler listenFunc }

func (mock *mockListen) Use(cb listenFunc) { mock.handler = cb }

type proxyServerAPI struct {
	mock *mockProxyServer
}

type mockProxyServer struct {
	api    proxyServerAPI
	Listen mockListen
}

func (api *proxyServerAPI) Listen(req *proxy.ListenRequest, srv proxy.API_ListenServer) error {
	if api.mock.Listen.handler != nil {
		return api.mock.Listen.handler(req, srv)
	}
	return errors.Errorf("unhandled pachd mock proxy.Listen")
}

// MockPachd provides an interface for running the interface for a Pachd API
// server locally without any of its dependencies. Tests may mock out specific
// API calls by providing a handler function, and later check information about
// the mocked calls.
type MockPachd struct {
	cancel  context.CancelFunc
	errchan chan error

	Addr net.Addr

	PFS           mockPFSServer
	PPS           mockPPSServer
	Auth          mockAuthServer
	GetAuthServer func() authserver.APIServer
	Transaction   mockTransactionServer
	Identity      mockIdentityServer
	Enterprise    mockEnterpriseServer
	License       mockLicenseServer
	Version       mockVersionServer
	Admin         mockAdminServer
	Proxy         mockProxyServer
}

type InterceptorOption func(mock *MockPachd) grpcutil.Interceptor

func AuthMiddlewareInterceptor(mock *MockPachd) grpcutil.Interceptor {
	return grpcutil.Interceptor{
		UnaryServerInterceptor:  authmw.NewInterceptor(mock.GetAuthServer).InterceptUnary,
		StreamServerInterceptor: authmw.NewInterceptor(mock.GetAuthServer).InterceptStream,
	}
}

// NewMockPachd constructs a mock Pachd API server whose behavior can be
// controlled through the MockPachd instance. By default, all API calls will
// error, unless a handler is specified.
// A port value of 0 will choose a free port automatically
// options can be supplied to configure the grpcutil.Server's interceptors.
func NewMockPachd(ctx context.Context, port uint16, options ...InterceptorOption) (*MockPachd, error) {
	mock := &MockPachd{
		errchan: make(chan error),
	}

	ctx, mock.cancel = context.WithCancel(ctx)

	mock.PFS.api.mock = &mock.PFS
	mock.PPS.api.mock = &mock.PPS
	mock.Auth.api.mock = &mock.Auth
	mock.Transaction.api.mock = &mock.Transaction
	mock.Enterprise.api.mock = &mock.Enterprise
	mock.License.api.mock = &mock.License
	mock.Version.api.mock = &mock.Version
	mock.Admin.api.mock = &mock.Admin
	mock.Proxy.api.mock = &mock.Proxy
	mock.Identity.api.mock = &mock.Identity
	mock.GetAuthServer = func() authserver.APIServer {
		return &mock.Auth.api
	}

	loggingInterceptor := loggingmw.NewLoggingInterceptor(logrus.StandardLogger())
	unaryOpts := []grpc.UnaryServerInterceptor{
		errorsmw.UnaryServerInterceptor,
		loggingInterceptor.UnaryServerInterceptor,
	}
	streamOpts := []grpc.StreamServerInterceptor{
		errorsmw.StreamServerInterceptor,
		loggingInterceptor.StreamServerInterceptor,
	}
	for _, opt := range options {
		interceptor := opt(mock)
		if interceptor.UnaryServerInterceptor != nil {
			unaryOpts = append(unaryOpts, interceptor.UnaryServerInterceptor)
		}
		if interceptor.StreamServerInterceptor != nil {
			streamOpts = append(streamOpts, interceptor.StreamServerInterceptor)
		}
	}
	server, err := grpcutil.NewServer(ctx, false,
		grpc.ChainUnaryInterceptor(
			unaryOpts...,
		),
		grpc.ChainStreamInterceptor(
			streamOpts...,
		),
	)
	if err != nil {
		return nil, err
	}

	admin.RegisterAPIServer(server.Server, &mock.Admin.api)
	auth.RegisterAPIServer(server.Server, &mock.Auth.api)
	enterprise.RegisterAPIServer(server.Server, &mock.Enterprise.api)
	pfs.RegisterAPIServer(server.Server, &mock.PFS.api)
	pps.RegisterAPIServer(server.Server, &mock.PPS.api)
	transaction.RegisterAPIServer(server.Server, &mock.Transaction.api)
	version.RegisterAPIServer(server.Server, &mock.Version.api)
	proxy.RegisterAPIServer(server.Server, &mock.Proxy.api)
	license.RegisterAPIServer(server.Server, &mock.License.api)
	identity.RegisterAPIServer(server.Server, &mock.Identity.api)

	listener, err := server.ListenTCP("localhost", port)
	if err != nil {
		return nil, err
	}

	go func() {
		mock.errchan <- server.Wait()
		close(mock.errchan)
	}()

	mock.Addr = listener.Addr()

	return mock, nil
}

// Err returns a read-only channel that will receive the first error that occurs
// in the server group (stopping all the servers).
func (mock *MockPachd) Err() <-chan error {
	return mock.errchan
}

// Close will cancel the mock Pachd API server goroutine and return its result
func (mock *MockPachd) Close() error {
	mock.cancel()
	return <-mock.errchan
}
