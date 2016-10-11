// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

// Package workspaces provides a client for Amazon WorkSpaces.
package workspaces

import (
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/service"
)

const opCreateWorkspaces = "CreateWorkspaces"

// CreateWorkspacesRequest generates a request for the CreateWorkspaces operation.
func (c *WorkSpaces) CreateWorkspacesRequest(input *CreateWorkspacesInput) (req *service.Request, output *CreateWorkspacesOutput) {
	op := &service.Operation{
		Name:       opCreateWorkspaces,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &CreateWorkspacesInput{}
	}

	req = c.newRequest(op, input, output)
	output = &CreateWorkspacesOutput{}
	req.Data = output
	return
}

// Creates one or more WorkSpaces.
//
//  This operation is asynchronous and returns before the WorkSpaces are created.
func (c *WorkSpaces) CreateWorkspaces(input *CreateWorkspacesInput) (*CreateWorkspacesOutput, error) {
	req, out := c.CreateWorkspacesRequest(input)
	err := req.Send()
	return out, err
}

const opDescribeWorkspaceBundles = "DescribeWorkspaceBundles"

// DescribeWorkspaceBundlesRequest generates a request for the DescribeWorkspaceBundles operation.
func (c *WorkSpaces) DescribeWorkspaceBundlesRequest(input *DescribeWorkspaceBundlesInput) (req *service.Request, output *DescribeWorkspaceBundlesOutput) {
	op := &service.Operation{
		Name:       opDescribeWorkspaceBundles,
		HTTPMethod: "POST",
		HTTPPath:   "/",
		Paginator: &service.Paginator{
			InputTokens:     []string{"NextToken"},
			OutputTokens:    []string{"NextToken"},
			LimitToken:      "",
			TruncationToken: "",
		},
	}

	if input == nil {
		input = &DescribeWorkspaceBundlesInput{}
	}

	req = c.newRequest(op, input, output)
	output = &DescribeWorkspaceBundlesOutput{}
	req.Data = output
	return
}

// Obtains information about the WorkSpace bundles that are available to your
// account in the specified region.
//
// You can filter the results with either the BundleIds parameter, or the Owner
// parameter, but not both.
//
// This operation supports pagination with the use of the NextToken request
// and response parameters. If more results are available, the NextToken response
// member contains a token that you pass in the next call to this operation
// to retrieve the next set of items.
func (c *WorkSpaces) DescribeWorkspaceBundles(input *DescribeWorkspaceBundlesInput) (*DescribeWorkspaceBundlesOutput, error) {
	req, out := c.DescribeWorkspaceBundlesRequest(input)
	err := req.Send()
	return out, err
}

func (c *WorkSpaces) DescribeWorkspaceBundlesPages(input *DescribeWorkspaceBundlesInput, fn func(p *DescribeWorkspaceBundlesOutput, lastPage bool) (shouldContinue bool)) error {
	page, _ := c.DescribeWorkspaceBundlesRequest(input)
	return page.EachPage(func(p interface{}, lastPage bool) bool {
		return fn(p.(*DescribeWorkspaceBundlesOutput), lastPage)
	})
}

const opDescribeWorkspaceDirectories = "DescribeWorkspaceDirectories"

// DescribeWorkspaceDirectoriesRequest generates a request for the DescribeWorkspaceDirectories operation.
func (c *WorkSpaces) DescribeWorkspaceDirectoriesRequest(input *DescribeWorkspaceDirectoriesInput) (req *service.Request, output *DescribeWorkspaceDirectoriesOutput) {
	op := &service.Operation{
		Name:       opDescribeWorkspaceDirectories,
		HTTPMethod: "POST",
		HTTPPath:   "/",
		Paginator: &service.Paginator{
			InputTokens:     []string{"NextToken"},
			OutputTokens:    []string{"NextToken"},
			LimitToken:      "",
			TruncationToken: "",
		},
	}

	if input == nil {
		input = &DescribeWorkspaceDirectoriesInput{}
	}

	req = c.newRequest(op, input, output)
	output = &DescribeWorkspaceDirectoriesOutput{}
	req.Data = output
	return
}

// Retrieves information about the AWS Directory Service directories in the
// region that are registered with Amazon WorkSpaces and are available to your
// account.
//
// This operation supports pagination with the use of the NextToken request
// and response parameters. If more results are available, the NextToken response
// member contains a token that you pass in the next call to this operation
// to retrieve the next set of items.
func (c *WorkSpaces) DescribeWorkspaceDirectories(input *DescribeWorkspaceDirectoriesInput) (*DescribeWorkspaceDirectoriesOutput, error) {
	req, out := c.DescribeWorkspaceDirectoriesRequest(input)
	err := req.Send()
	return out, err
}

func (c *WorkSpaces) DescribeWorkspaceDirectoriesPages(input *DescribeWorkspaceDirectoriesInput, fn func(p *DescribeWorkspaceDirectoriesOutput, lastPage bool) (shouldContinue bool)) error {
	page, _ := c.DescribeWorkspaceDirectoriesRequest(input)
	return page.EachPage(func(p interface{}, lastPage bool) bool {
		return fn(p.(*DescribeWorkspaceDirectoriesOutput), lastPage)
	})
}

const opDescribeWorkspaces = "DescribeWorkspaces"

// DescribeWorkspacesRequest generates a request for the DescribeWorkspaces operation.
func (c *WorkSpaces) DescribeWorkspacesRequest(input *DescribeWorkspacesInput) (req *service.Request, output *DescribeWorkspacesOutput) {
	op := &service.Operation{
		Name:       opDescribeWorkspaces,
		HTTPMethod: "POST",
		HTTPPath:   "/",
		Paginator: &service.Paginator{
			InputTokens:     []string{"NextToken"},
			OutputTokens:    []string{"NextToken"},
			LimitToken:      "Limit",
			TruncationToken: "",
		},
	}

	if input == nil {
		input = &DescribeWorkspacesInput{}
	}

	req = c.newRequest(op, input, output)
	output = &DescribeWorkspacesOutput{}
	req.Data = output
	return
}

// Obtains information about the specified WorkSpaces.
//
// Only one of the filter parameters, such as BundleId, DirectoryId, or WorkspaceIds,
// can be specified at a time.
//
// This operation supports pagination with the use of the NextToken request
// and response parameters. If more results are available, the NextToken response
// member contains a token that you pass in the next call to this operation
// to retrieve the next set of items.
func (c *WorkSpaces) DescribeWorkspaces(input *DescribeWorkspacesInput) (*DescribeWorkspacesOutput, error) {
	req, out := c.DescribeWorkspacesRequest(input)
	err := req.Send()
	return out, err
}

func (c *WorkSpaces) DescribeWorkspacesPages(input *DescribeWorkspacesInput, fn func(p *DescribeWorkspacesOutput, lastPage bool) (shouldContinue bool)) error {
	page, _ := c.DescribeWorkspacesRequest(input)
	return page.EachPage(func(p interface{}, lastPage bool) bool {
		return fn(p.(*DescribeWorkspacesOutput), lastPage)
	})
}

const opRebootWorkspaces = "RebootWorkspaces"

// RebootWorkspacesRequest generates a request for the RebootWorkspaces operation.
func (c *WorkSpaces) RebootWorkspacesRequest(input *RebootWorkspacesInput) (req *service.Request, output *RebootWorkspacesOutput) {
	op := &service.Operation{
		Name:       opRebootWorkspaces,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &RebootWorkspacesInput{}
	}

	req = c.newRequest(op, input, output)
	output = &RebootWorkspacesOutput{}
	req.Data = output
	return
}

// Reboots the specified WorkSpaces.
//
// To be able to reboot a WorkSpace, the WorkSpace must have a State of AVAILABLE,
// IMPAIRED, or INOPERABLE.
//
//  This operation is asynchronous and will return before the WorkSpaces have
// rebooted.
func (c *WorkSpaces) RebootWorkspaces(input *RebootWorkspacesInput) (*RebootWorkspacesOutput, error) {
	req, out := c.RebootWorkspacesRequest(input)
	err := req.Send()
	return out, err
}

const opRebuildWorkspaces = "RebuildWorkspaces"

// RebuildWorkspacesRequest generates a request for the RebuildWorkspaces operation.
func (c *WorkSpaces) RebuildWorkspacesRequest(input *RebuildWorkspacesInput) (req *service.Request, output *RebuildWorkspacesOutput) {
	op := &service.Operation{
		Name:       opRebuildWorkspaces,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &RebuildWorkspacesInput{}
	}

	req = c.newRequest(op, input, output)
	output = &RebuildWorkspacesOutput{}
	req.Data = output
	return
}

// Rebuilds the specified WorkSpaces.
//
// Rebuilding a WorkSpace is a potentially destructive action that can result
// in the loss of data. Rebuilding a WorkSpace causes the following to occur:
//
//  The system is restored to the image of the bundle that the WorkSpace is
// created from. Any applications that have been installed, or system settings
// that have been made since the WorkSpace was created will be lost. The data
// drive (D drive) is re-created from the last automatic snapshot taken of the
// data drive. The current contents of the data drive are overwritten. Automatic
// snapshots of the data drive are taken every 12 hours, so the snapshot can
// be as much as 12 hours old.  To be able to rebuild a WorkSpace, the WorkSpace
// must have a State of AVAILABLE or ERROR.
//
//  This operation is asynchronous and will return before the WorkSpaces have
// been completely rebuilt.
func (c *WorkSpaces) RebuildWorkspaces(input *RebuildWorkspacesInput) (*RebuildWorkspacesOutput, error) {
	req, out := c.RebuildWorkspacesRequest(input)
	err := req.Send()
	return out, err
}

const opTerminateWorkspaces = "TerminateWorkspaces"

// TerminateWorkspacesRequest generates a request for the TerminateWorkspaces operation.
func (c *WorkSpaces) TerminateWorkspacesRequest(input *TerminateWorkspacesInput) (req *service.Request, output *TerminateWorkspacesOutput) {
	op := &service.Operation{
		Name:       opTerminateWorkspaces,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &TerminateWorkspacesInput{}
	}

	req = c.newRequest(op, input, output)
	output = &TerminateWorkspacesOutput{}
	req.Data = output
	return
}

// Terminates the specified WorkSpaces.
//
// Terminating a WorkSpace is a permanent action and cannot be undone. The
// user's data is not maintained and will be destroyed. If you need to archive
// any user data, contact Amazon Web Services before terminating the WorkSpace.
//
// You can terminate a WorkSpace that is in any state except SUSPENDED.
//
//  This operation is asynchronous and will return before the WorkSpaces have
// been completely terminated.
func (c *WorkSpaces) TerminateWorkspaces(input *TerminateWorkspacesInput) (*TerminateWorkspacesOutput, error) {
	req, out := c.TerminateWorkspacesRequest(input)
	err := req.Send()
	return out, err
}

// Contains information about the compute type of a WorkSpace bundle.
type ComputeType struct {
	// The name of the compute type for the bundle.
	Name *string `type:"string" enum:"Compute"`

	metadataComputeType `json:"-" xml:"-"`
}

type metadataComputeType struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s ComputeType) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s ComputeType) GoString() string {
	return s.String()
}

// Contains the inputs for the CreateWorkspaces operation.
type CreateWorkspacesInput struct {
	// An array of structures that specify the WorkSpaces to create.
	Workspaces []*WorkspaceRequest `type:"list" required:"true"`

	metadataCreateWorkspacesInput `json:"-" xml:"-"`
}

type metadataCreateWorkspacesInput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s CreateWorkspacesInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s CreateWorkspacesInput) GoString() string {
	return s.String()
}

// Contains the result of the CreateWorkspaces operation.
type CreateWorkspacesOutput struct {
	// An array of structures that represent the WorkSpaces that could not be created.
	FailedRequests []*FailedCreateWorkspaceRequest `type:"list"`

	// An array of structures that represent the WorkSpaces that were created.
	//
	// Because this operation is asynchronous, the identifier in WorkspaceId is
	// not immediately available. If you immediately call DescribeWorkspaces with
	// this identifier, no information will be returned.
	PendingRequests []*Workspace `type:"list"`

	metadataCreateWorkspacesOutput `json:"-" xml:"-"`
}

type metadataCreateWorkspacesOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s CreateWorkspacesOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s CreateWorkspacesOutput) GoString() string {
	return s.String()
}

// Contains default WorkSpace creation information.
type DefaultWorkspaceCreationProperties struct {
	// The identifier of any custom security groups that are applied to the WorkSpaces
	// when they are created.
	CustomSecurityGroupID *string `locationName:"CustomSecurityGroupId" type:"string"`

	// The organizational unit (OU) in the directory that the WorkSpace machine
	// accounts are placed in.
	DefaultOU *string `locationName:"DefaultOu" type:"string"`

	// A public IP address will be attached to all WorkSpaces that are created or
	// rebuilt.
	EnableInternetAccess *bool `type:"boolean"`

	// Specifies if the directory is enabled for Amazon WorkDocs.
	EnableWorkDocs *bool `type:"boolean"`

	// The WorkSpace user is an administrator on the WorkSpace.
	UserEnabledAsLocalAdministrator *bool `type:"boolean"`

	metadataDefaultWorkspaceCreationProperties `json:"-" xml:"-"`
}

type metadataDefaultWorkspaceCreationProperties struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s DefaultWorkspaceCreationProperties) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DefaultWorkspaceCreationProperties) GoString() string {
	return s.String()
}

// Contains the inputs for the DescribeWorkspaceBundles operation.
type DescribeWorkspaceBundlesInput struct {
	// An array of strings that contains the identifiers of the bundles to retrieve.
	// This parameter cannot be combined with any other filter parameter.
	BundleIDs []*string `locationName:"BundleIds" type:"list"`

	// The NextToken value from a previous call to this operation. Pass null if
	// this is the first call.
	NextToken *string `type:"string"`

	// The owner of the bundles to retrieve. This parameter cannot be combined with
	// any other filter parameter.
	//
	// This contains one of the following values:
	//
	//  null - Retrieves the bundles that belong to the account making the call.
	//  AMAZON - Retrieves the bundles that are provided by AWS.
	Owner *string `type:"string"`

	metadataDescribeWorkspaceBundlesInput `json:"-" xml:"-"`
}

type metadataDescribeWorkspaceBundlesInput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s DescribeWorkspaceBundlesInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DescribeWorkspaceBundlesInput) GoString() string {
	return s.String()
}

// Contains the results of the DescribeWorkspaceBundles operation.
type DescribeWorkspaceBundlesOutput struct {
	// An array of structures that contain information about the bundles.
	Bundles []*WorkspaceBundle `type:"list"`

	// If not null, more results are available. Pass this value for the NextToken
	// parameter in a subsequent call to this operation to retrieve the next set
	// of items. This token is valid for one day and must be used within that timeframe.
	NextToken *string `type:"string"`

	metadataDescribeWorkspaceBundlesOutput `json:"-" xml:"-"`
}

type metadataDescribeWorkspaceBundlesOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s DescribeWorkspaceBundlesOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DescribeWorkspaceBundlesOutput) GoString() string {
	return s.String()
}

// Contains the inputs for the DescribeWorkspaceDirectories operation.
type DescribeWorkspaceDirectoriesInput struct {
	// An array of strings that contains the directory identifiers to retrieve information
	// for. If this member is null, all directories are retrieved.
	DirectoryIDs []*string `locationName:"DirectoryIds" type:"list"`

	// The NextToken value from a previous call to this operation. Pass null if
	// this is the first call.
	NextToken *string `type:"string"`

	metadataDescribeWorkspaceDirectoriesInput `json:"-" xml:"-"`
}

type metadataDescribeWorkspaceDirectoriesInput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s DescribeWorkspaceDirectoriesInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DescribeWorkspaceDirectoriesInput) GoString() string {
	return s.String()
}

// Contains the results of the DescribeWorkspaceDirectories operation.
type DescribeWorkspaceDirectoriesOutput struct {
	// An array of structures that contain information about the directories.
	Directories []*WorkspaceDirectory `type:"list"`

	// If not null, more results are available. Pass this value for the NextToken
	// parameter in a subsequent call to this operation to retrieve the next set
	// of items. This token is valid for one day and must be used within that timeframe.
	NextToken *string `type:"string"`

	metadataDescribeWorkspaceDirectoriesOutput `json:"-" xml:"-"`
}

type metadataDescribeWorkspaceDirectoriesOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s DescribeWorkspaceDirectoriesOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DescribeWorkspaceDirectoriesOutput) GoString() string {
	return s.String()
}

// Contains the inputs for the DescribeWorkspaces operation.
type DescribeWorkspacesInput struct {
	// The identifier of a bundle to obtain the WorkSpaces for. All WorkSpaces that
	// are created from this bundle will be retrieved. This parameter cannot be
	// combined with any other filter parameter.
	BundleID *string `locationName:"BundleId" type:"string"`

	// Specifies the directory identifier to which to limit the WorkSpaces. Optionally,
	// you can specify a specific directory user with the UserName parameter. This
	// parameter cannot be combined with any other filter parameter.
	DirectoryID *string `locationName:"DirectoryId" type:"string"`

	// The maximum number of items to return.
	Limit *int64 `type:"integer"`

	// The NextToken value from a previous call to this operation. Pass null if
	// this is the first call.
	NextToken *string `type:"string"`

	// Used with the DirectoryId parameter to specify the directory user for which
	// to obtain the WorkSpace.
	UserName *string `type:"string"`

	// An array of strings that contain the identifiers of the WorkSpaces for which
	// to retrieve information. This parameter cannot be combined with any other
	// filter parameter.
	//
	// Because the CreateWorkspaces operation is asynchronous, the identifier returned
	// by CreateWorkspaces is not immediately available. If you immediately call
	// DescribeWorkspaces with this identifier, no information will be returned.
	WorkspaceIDs []*string `locationName:"WorkspaceIds" type:"list"`

	metadataDescribeWorkspacesInput `json:"-" xml:"-"`
}

type metadataDescribeWorkspacesInput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s DescribeWorkspacesInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DescribeWorkspacesInput) GoString() string {
	return s.String()
}

// Contains the results for the DescribeWorkspaces operation.
type DescribeWorkspacesOutput struct {
	// If not null, more results are available. Pass this value for the NextToken
	// parameter in a subsequent call to this operation to retrieve the next set
	// of items. This token is valid for one day and must be used within that timeframe.
	NextToken *string `type:"string"`

	// An array of structures that contain the information about the WorkSpaces.
	//
	// Because the CreateWorkspaces operation is asynchronous, some of this information
	// may be incomplete for a newly-created WorkSpace.
	Workspaces []*Workspace `type:"list"`

	metadataDescribeWorkspacesOutput `json:"-" xml:"-"`
}

type metadataDescribeWorkspacesOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s DescribeWorkspacesOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s DescribeWorkspacesOutput) GoString() string {
	return s.String()
}

// Contains information about a WorkSpace that could not be created.
type FailedCreateWorkspaceRequest struct {
	// The error code.
	ErrorCode *string `type:"string"`

	// The textual error message.
	ErrorMessage *string `type:"string"`

	// A WorkspaceRequest object that contains the information about the WorkSpace
	// that could not be created.
	WorkspaceRequest *WorkspaceRequest `type:"structure"`

	metadataFailedCreateWorkspaceRequest `json:"-" xml:"-"`
}

type metadataFailedCreateWorkspaceRequest struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s FailedCreateWorkspaceRequest) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s FailedCreateWorkspaceRequest) GoString() string {
	return s.String()
}

// Contains information about a WorkSpace that could not be rebooted (RebootWorkspaces),
// rebuilt (RebuildWorkspaces), or terminated (TerminateWorkspaces).
type FailedWorkspaceChangeRequest struct {
	// The error code.
	ErrorCode *string `type:"string"`

	// The textual error message.
	ErrorMessage *string `type:"string"`

	// The identifier of the WorkSpace.
	WorkspaceID *string `locationName:"WorkspaceId" type:"string"`

	metadataFailedWorkspaceChangeRequest `json:"-" xml:"-"`
}

type metadataFailedWorkspaceChangeRequest struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s FailedWorkspaceChangeRequest) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s FailedWorkspaceChangeRequest) GoString() string {
	return s.String()
}

// Contains information used with the RebootWorkspaces operation to reboot a
// WorkSpace.
type RebootRequest struct {
	// The identifier of the WorkSpace to reboot.
	WorkspaceID *string `locationName:"WorkspaceId" type:"string" required:"true"`

	metadataRebootRequest `json:"-" xml:"-"`
}

type metadataRebootRequest struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s RebootRequest) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s RebootRequest) GoString() string {
	return s.String()
}

// Contains the inputs for the RebootWorkspaces operation.
type RebootWorkspacesInput struct {
	// An array of structures that specify the WorkSpaces to reboot.
	RebootWorkspaceRequests []*RebootRequest `type:"list" required:"true"`

	metadataRebootWorkspacesInput `json:"-" xml:"-"`
}

type metadataRebootWorkspacesInput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s RebootWorkspacesInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s RebootWorkspacesInput) GoString() string {
	return s.String()
}

// Contains the results of the RebootWorkspaces operation.
type RebootWorkspacesOutput struct {
	// An array of structures that represent any WorkSpaces that could not be rebooted.
	FailedRequests []*FailedWorkspaceChangeRequest `type:"list"`

	metadataRebootWorkspacesOutput `json:"-" xml:"-"`
}

type metadataRebootWorkspacesOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s RebootWorkspacesOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s RebootWorkspacesOutput) GoString() string {
	return s.String()
}

// Contains information used with the RebuildWorkspaces operation to rebuild
// a WorkSpace.
type RebuildRequest struct {
	// The identifier of the WorkSpace to rebuild.
	WorkspaceID *string `locationName:"WorkspaceId" type:"string" required:"true"`

	metadataRebuildRequest `json:"-" xml:"-"`
}

type metadataRebuildRequest struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s RebuildRequest) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s RebuildRequest) GoString() string {
	return s.String()
}

// Contains the inputs for the RebuildWorkspaces operation.
type RebuildWorkspacesInput struct {
	// An array of structures that specify the WorkSpaces to rebuild.
	RebuildWorkspaceRequests []*RebuildRequest `type:"list" required:"true"`

	metadataRebuildWorkspacesInput `json:"-" xml:"-"`
}

type metadataRebuildWorkspacesInput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s RebuildWorkspacesInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s RebuildWorkspacesInput) GoString() string {
	return s.String()
}

// Contains the results of the RebuildWorkspaces operation.
type RebuildWorkspacesOutput struct {
	// An array of structures that represent any WorkSpaces that could not be rebuilt.
	FailedRequests []*FailedWorkspaceChangeRequest `type:"list"`

	metadataRebuildWorkspacesOutput `json:"-" xml:"-"`
}

type metadataRebuildWorkspacesOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s RebuildWorkspacesOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s RebuildWorkspacesOutput) GoString() string {
	return s.String()
}

// Contains information used with the TerminateWorkspaces operation to terminate
// a WorkSpace.
type TerminateRequest struct {
	// The identifier of the WorkSpace to terminate.
	WorkspaceID *string `locationName:"WorkspaceId" type:"string" required:"true"`

	metadataTerminateRequest `json:"-" xml:"-"`
}

type metadataTerminateRequest struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s TerminateRequest) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s TerminateRequest) GoString() string {
	return s.String()
}

// Contains the inputs for the TerminateWorkspaces operation.
type TerminateWorkspacesInput struct {
	// An array of structures that specify the WorkSpaces to terminate.
	TerminateWorkspaceRequests []*TerminateRequest `type:"list" required:"true"`

	metadataTerminateWorkspacesInput `json:"-" xml:"-"`
}

type metadataTerminateWorkspacesInput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s TerminateWorkspacesInput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s TerminateWorkspacesInput) GoString() string {
	return s.String()
}

// Contains the results of the TerminateWorkspaces operation.
type TerminateWorkspacesOutput struct {
	// An array of structures that represent any WorkSpaces that could not be terminated.
	FailedRequests []*FailedWorkspaceChangeRequest `type:"list"`

	metadataTerminateWorkspacesOutput `json:"-" xml:"-"`
}

type metadataTerminateWorkspacesOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s TerminateWorkspacesOutput) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s TerminateWorkspacesOutput) GoString() string {
	return s.String()
}

// Contains information about the user storage for a WorkSpace bundle.
type UserStorage struct {
	// The amount of user storage for the bundle.
	Capacity *string `type:"string"`

	metadataUserStorage `json:"-" xml:"-"`
}

type metadataUserStorage struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s UserStorage) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s UserStorage) GoString() string {
	return s.String()
}

// Contains information about a WorkSpace.
type Workspace struct {
	// The identifier of the bundle that the WorkSpace was created from.
	BundleID *string `locationName:"BundleId" type:"string"`

	// The identifier of the AWS Directory Service directory that the WorkSpace
	// belongs to.
	DirectoryID *string `locationName:"DirectoryId" type:"string"`

	// If the WorkSpace could not be created, this contains the error code.
	ErrorCode *string `type:"string"`

	// If the WorkSpace could not be created, this contains a textual error message
	// that describes the failure.
	ErrorMessage *string `type:"string"`

	// The IP address of the WorkSpace.
	IPAddress *string `locationName:"IpAddress" type:"string"`

	// The operational state of the WorkSpace.
	State *string `type:"string" enum:"WorkspaceState"`

	// The identifier of the subnet that the WorkSpace is in.
	SubnetID *string `locationName:"SubnetId" type:"string"`

	// The user that the WorkSpace is assigned to.
	UserName *string `type:"string"`

	// The identifier of the WorkSpace.
	WorkspaceID *string `locationName:"WorkspaceId" type:"string"`

	metadataWorkspace `json:"-" xml:"-"`
}

type metadataWorkspace struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s Workspace) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s Workspace) GoString() string {
	return s.String()
}

// Contains information about a WorkSpace bundle.
type WorkspaceBundle struct {
	// The bundle identifier.
	BundleID *string `locationName:"BundleId" type:"string"`

	// A ComputeType object that specifies the compute type for the bundle.
	ComputeType *ComputeType `type:"structure"`

	// The bundle description.
	Description *string `type:"string"`

	// The name of the bundle.
	Name *string `type:"string"`

	// The owner of the bundle. This contains the owner's account identifier, or
	// AMAZON if the bundle is provided by AWS.
	Owner *string `type:"string"`

	// A UserStorage object that specifies the amount of user storage that the bundle
	// contains.
	UserStorage *UserStorage `type:"structure"`

	metadataWorkspaceBundle `json:"-" xml:"-"`
}

type metadataWorkspaceBundle struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s WorkspaceBundle) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s WorkspaceBundle) GoString() string {
	return s.String()
}

// Contains information about an AWS Directory Service directory for use with
// Amazon WorkSpaces.
type WorkspaceDirectory struct {
	// The directory alias.
	Alias *string `type:"string"`

	// The user name for the service account.
	CustomerUserName *string `type:"string"`

	// An array of strings that contains the IP addresses of the DNS servers for
	// the directory.
	DNSIPAddresses []*string `locationName:"DnsIpAddresses" type:"list"`

	// The directory identifier.
	DirectoryID *string `locationName:"DirectoryId" type:"string"`

	// The name of the directory.
	DirectoryName *string `type:"string"`

	// The directory type.
	DirectoryType *string `type:"string" enum:"WorkspaceDirectoryType"`

	// The identifier of the IAM role. This is the role that allows Amazon WorkSpaces
	// to make calls to other services, such as Amazon EC2, on your behalf.
	IAMRoleID *string `locationName:"IamRoleId" type:"string"`

	// The registration code for the directory. This is the code that users enter
	// in their Amazon WorkSpaces client application to connect to the directory.
	RegistrationCode *string `type:"string"`

	// The state of the directory's registration with Amazon WorkSpaces
	State *string `type:"string" enum:"WorkspaceDirectoryState"`

	// An array of strings that contains the identifiers of the subnets used with
	// the directory.
	SubnetIDs []*string `locationName:"SubnetIds" type:"list"`

	// A structure that specifies the default creation properties for all WorkSpaces
	// in the directory.
	WorkspaceCreationProperties *DefaultWorkspaceCreationProperties `type:"structure"`

	// The identifier of the security group that is assigned to new WorkSpaces.
	WorkspaceSecurityGroupID *string `locationName:"WorkspaceSecurityGroupId" type:"string"`

	metadataWorkspaceDirectory `json:"-" xml:"-"`
}

type metadataWorkspaceDirectory struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s WorkspaceDirectory) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s WorkspaceDirectory) GoString() string {
	return s.String()
}

// Contains information about a WorkSpace creation request.
type WorkspaceRequest struct {
	// The identifier of the bundle to create the WorkSpace from. You can use the
	// DescribeWorkspaceBundles operation to obtain a list of the bundles that are
	// available.
	BundleID *string `locationName:"BundleId" type:"string" required:"true"`

	// The identifier of the AWS Directory Service directory to create the WorkSpace
	// in. You can use the DescribeWorkspaceDirectories operation to obtain a list
	// of the directories that are available.
	DirectoryID *string `locationName:"DirectoryId" type:"string" required:"true"`

	// The username that the WorkSpace is assigned to. This username must exist
	// in the AWS Directory Service directory specified by the DirectoryId member.
	UserName *string `type:"string" required:"true"`

	metadataWorkspaceRequest `json:"-" xml:"-"`
}

type metadataWorkspaceRequest struct {
	SDKShapeTraits bool `type:"structure"`
}

// String returns the string representation
func (s WorkspaceRequest) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s WorkspaceRequest) GoString() string {
	return s.String()
}

const (
	// @enum Compute
	ComputeValue = "VALUE"
	// @enum Compute
	ComputeStandard = "STANDARD"
	// @enum Compute
	ComputePerformance = "PERFORMANCE"
)

const (
	// @enum WorkspaceDirectoryState
	WorkspaceDirectoryStateRegistering = "REGISTERING"
	// @enum WorkspaceDirectoryState
	WorkspaceDirectoryStateRegistered = "REGISTERED"
	// @enum WorkspaceDirectoryState
	WorkspaceDirectoryStateDeregistering = "DEREGISTERING"
	// @enum WorkspaceDirectoryState
	WorkspaceDirectoryStateDeregistered = "DEREGISTERED"
	// @enum WorkspaceDirectoryState
	WorkspaceDirectoryStateError = "ERROR"
)

const (
	// @enum WorkspaceDirectoryType
	WorkspaceDirectoryTypeSimpleAd = "SIMPLE_AD"
	// @enum WorkspaceDirectoryType
	WorkspaceDirectoryTypeAdConnector = "AD_CONNECTOR"
)

const (
	// @enum WorkspaceState
	WorkspaceStatePending = "PENDING"
	// @enum WorkspaceState
	WorkspaceStateAvailable = "AVAILABLE"
	// @enum WorkspaceState
	WorkspaceStateImpaired = "IMPAIRED"
	// @enum WorkspaceState
	WorkspaceStateUnhealthy = "UNHEALTHY"
	// @enum WorkspaceState
	WorkspaceStateRebooting = "REBOOTING"
	// @enum WorkspaceState
	WorkspaceStateRebuilding = "REBUILDING"
	// @enum WorkspaceState
	WorkspaceStateTerminating = "TERMINATING"
	// @enum WorkspaceState
	WorkspaceStateTerminated = "TERMINATED"
	// @enum WorkspaceState
	WorkspaceStateSuspended = "SUSPENDED"
	// @enum WorkspaceState
	WorkspaceStateError = "ERROR"
)
