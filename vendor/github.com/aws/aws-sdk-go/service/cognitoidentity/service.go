// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

package cognitoidentity

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/private/protocol/jsonrpc"
)

// CognitoIdentity provides the API operation methods for making requests to
// Amazon Cognito Identity. See this package's package overview docs
// for details on the service.
//
// CognitoIdentity methods are safe to use concurrently. It is not safe to
// modify mutate any of the struct's properties though.
type CognitoIdentity struct {
	*client.Client
}

// Used for custom client initialization logic
var initClient func(*client.Client)

// Used for custom request initialization logic
var initRequest func(*request.Request)

// Service information constants
const (
	ServiceName = "cognito-identity" // Name of service.
	EndpointsID = ServiceName        // ID to lookup a service endpoint with.
	ServiceID   = "Cognito Identity" // ServiceID is a unique identifer of a specific service.
)

// New creates a new instance of the CognitoIdentity client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a CognitoIdentity client from just a session.
//     svc := cognitoidentity.New(mySession)
//
//     // Create a CognitoIdentity client with additional configuration
//     svc := cognitoidentity.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func New(p client.ConfigProvider, cfgs ...*aws.Config) *CognitoIdentity {
	c := p.ClientConfig(EndpointsID, cfgs...)
	return newClient(*c.Config, c.Handlers, c.PartitionID, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newClient(cfg aws.Config, handlers request.Handlers, partitionID, endpoint, signingRegion, signingName string) *CognitoIdentity {
	svc := &CognitoIdentity{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   ServiceName,
				ServiceID:     ServiceID,
				SigningName:   signingName,
				SigningRegion: signingRegion,
				PartitionID:   partitionID,
				Endpoint:      endpoint,
				APIVersion:    "2014-06-30",
				JSONVersion:   "1.1",
				TargetPrefix:  "AWSCognitoIdentityService",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(jsonrpc.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(jsonrpc.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(jsonrpc.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(jsonrpc.UnmarshalErrorHandler)

	// Run custom client initialization if present
	if initClient != nil {
		initClient(svc.Client)
	}

	return svc
}

// newRequest creates a new request for a CognitoIdentity operation and runs any
// custom request initialization.
func (c *CognitoIdentity) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	// Run custom request initialization if present
	if initRequest != nil {
		initRequest(req)
	}

	return req
}
