// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

package cloudsearch

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/service"
	"github.com/aws/aws-sdk-go/internal/protocol/query"
	"github.com/aws/aws-sdk-go/internal/signer/v4"
)

// You use the Amazon CloudSearch configuration service to create, configure,
// and manage search domains. Configuration service requests are submitted using
// the AWS Query protocol. AWS Query requests are HTTP or HTTPS requests submitted
// via HTTP GET or POST with a query parameter named Action.
//
// The endpoint for configuration service requests is region-specific: cloudsearch.region.amazonaws.com.
// For example, cloudsearch.us-east-1.amazonaws.com. For a current list of supported
// regions and endpoints, see Regions and Endpoints (http://docs.aws.amazon.com/general/latest/gr/rande.html#cloudsearch_region"
// target="_blank).
type CloudSearch struct {
	*service.Service
}

// Used for custom service initialization logic
var initService func(*service.Service)

// Used for custom request initialization logic
var initRequest func(*service.Request)

// New returns a new CloudSearch client.
func New(config *aws.Config) *CloudSearch {
	service := &service.Service{
		Config:      defaults.DefaultConfig.Merge(config),
		ServiceName: "cloudsearch",
		APIVersion:  "2013-01-01",
	}
	service.Initialize()

	// Handlers
	service.Handlers.Sign.PushBack(v4.Sign)
	service.Handlers.Build.PushBack(query.Build)
	service.Handlers.Unmarshal.PushBack(query.Unmarshal)
	service.Handlers.UnmarshalMeta.PushBack(query.UnmarshalMeta)
	service.Handlers.UnmarshalError.PushBack(query.UnmarshalError)

	// Run custom service initialization if present
	if initService != nil {
		initService(service)
	}

	return &CloudSearch{service}
}

// newRequest creates a new request for a CloudSearch operation and runs any
// custom request initialization.
func (c *CloudSearch) newRequest(op *service.Operation, params, data interface{}) *service.Request {
	req := service.NewRequest(c.Service, op, params, data)

	// Run custom request initialization if present
	if initRequest != nil {
		initRequest(req)
	}

	return req
}