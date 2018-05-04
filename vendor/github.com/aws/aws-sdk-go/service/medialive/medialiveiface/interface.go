// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

// Package medialiveiface provides an interface to enable mocking the AWS Elemental MediaLive service client
// for testing your code.
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters.
package medialiveiface

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/medialive"
)

// MediaLiveAPI provides an interface to enable mocking the
// medialive.MediaLive service client's API operation,
// paginators, and waiters. This make unit testing your code that calls out
// to the SDK's service client's calls easier.
//
// The best way to use this interface is so the SDK's service client's calls
// can be stubbed out for unit testing your code with the SDK without needing
// to inject custom request handlers into the SDK's request pipeline.
//
//    // myFunc uses an SDK service client to make a request to
//    // AWS Elemental MediaLive.
//    func myFunc(svc medialiveiface.MediaLiveAPI) bool {
//        // Make svc.CreateChannel request
//    }
//
//    func main() {
//        sess := session.New()
//        svc := medialive.New(sess)
//
//        myFunc(svc)
//    }
//
// In your _test.go file:
//
//    // Define a mock struct to be used in your unit tests of myFunc.
//    type mockMediaLiveClient struct {
//        medialiveiface.MediaLiveAPI
//    }
//    func (m *mockMediaLiveClient) CreateChannel(input *medialive.CreateChannelInput) (*medialive.CreateChannelOutput, error) {
//        // mock response/functionality
//    }
//
//    func TestMyFunc(t *testing.T) {
//        // Setup Test
//        mockSvc := &mockMediaLiveClient{}
//
//        myfunc(mockSvc)
//
//        // Verify myFunc's functionality
//    }
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters. Its suggested to use the pattern above for testing, or using
// tooling to generate mocks to satisfy the interfaces.
type MediaLiveAPI interface {
	CreateChannel(*medialive.CreateChannelInput) (*medialive.CreateChannelOutput, error)
	CreateChannelWithContext(aws.Context, *medialive.CreateChannelInput, ...request.Option) (*medialive.CreateChannelOutput, error)
	CreateChannelRequest(*medialive.CreateChannelInput) (*request.Request, *medialive.CreateChannelOutput)

	CreateInput(*medialive.CreateInputInput) (*medialive.CreateInputOutput, error)
	CreateInputWithContext(aws.Context, *medialive.CreateInputInput, ...request.Option) (*medialive.CreateInputOutput, error)
	CreateInputRequest(*medialive.CreateInputInput) (*request.Request, *medialive.CreateInputOutput)

	CreateInputSecurityGroup(*medialive.CreateInputSecurityGroupInput) (*medialive.CreateInputSecurityGroupOutput, error)
	CreateInputSecurityGroupWithContext(aws.Context, *medialive.CreateInputSecurityGroupInput, ...request.Option) (*medialive.CreateInputSecurityGroupOutput, error)
	CreateInputSecurityGroupRequest(*medialive.CreateInputSecurityGroupInput) (*request.Request, *medialive.CreateInputSecurityGroupOutput)

	DeleteChannel(*medialive.DeleteChannelInput) (*medialive.DeleteChannelOutput, error)
	DeleteChannelWithContext(aws.Context, *medialive.DeleteChannelInput, ...request.Option) (*medialive.DeleteChannelOutput, error)
	DeleteChannelRequest(*medialive.DeleteChannelInput) (*request.Request, *medialive.DeleteChannelOutput)

	DeleteInput(*medialive.DeleteInputInput) (*medialive.DeleteInputOutput, error)
	DeleteInputWithContext(aws.Context, *medialive.DeleteInputInput, ...request.Option) (*medialive.DeleteInputOutput, error)
	DeleteInputRequest(*medialive.DeleteInputInput) (*request.Request, *medialive.DeleteInputOutput)

	DeleteInputSecurityGroup(*medialive.DeleteInputSecurityGroupInput) (*medialive.DeleteInputSecurityGroupOutput, error)
	DeleteInputSecurityGroupWithContext(aws.Context, *medialive.DeleteInputSecurityGroupInput, ...request.Option) (*medialive.DeleteInputSecurityGroupOutput, error)
	DeleteInputSecurityGroupRequest(*medialive.DeleteInputSecurityGroupInput) (*request.Request, *medialive.DeleteInputSecurityGroupOutput)

	DescribeChannel(*medialive.DescribeChannelInput) (*medialive.DescribeChannelOutput, error)
	DescribeChannelWithContext(aws.Context, *medialive.DescribeChannelInput, ...request.Option) (*medialive.DescribeChannelOutput, error)
	DescribeChannelRequest(*medialive.DescribeChannelInput) (*request.Request, *medialive.DescribeChannelOutput)

	DescribeInput(*medialive.DescribeInputInput) (*medialive.DescribeInputOutput, error)
	DescribeInputWithContext(aws.Context, *medialive.DescribeInputInput, ...request.Option) (*medialive.DescribeInputOutput, error)
	DescribeInputRequest(*medialive.DescribeInputInput) (*request.Request, *medialive.DescribeInputOutput)

	DescribeInputSecurityGroup(*medialive.DescribeInputSecurityGroupInput) (*medialive.DescribeInputSecurityGroupOutput, error)
	DescribeInputSecurityGroupWithContext(aws.Context, *medialive.DescribeInputSecurityGroupInput, ...request.Option) (*medialive.DescribeInputSecurityGroupOutput, error)
	DescribeInputSecurityGroupRequest(*medialive.DescribeInputSecurityGroupInput) (*request.Request, *medialive.DescribeInputSecurityGroupOutput)

	ListChannels(*medialive.ListChannelsInput) (*medialive.ListChannelsOutput, error)
	ListChannelsWithContext(aws.Context, *medialive.ListChannelsInput, ...request.Option) (*medialive.ListChannelsOutput, error)
	ListChannelsRequest(*medialive.ListChannelsInput) (*request.Request, *medialive.ListChannelsOutput)

	ListChannelsPages(*medialive.ListChannelsInput, func(*medialive.ListChannelsOutput, bool) bool) error
	ListChannelsPagesWithContext(aws.Context, *medialive.ListChannelsInput, func(*medialive.ListChannelsOutput, bool) bool, ...request.Option) error

	ListInputSecurityGroups(*medialive.ListInputSecurityGroupsInput) (*medialive.ListInputSecurityGroupsOutput, error)
	ListInputSecurityGroupsWithContext(aws.Context, *medialive.ListInputSecurityGroupsInput, ...request.Option) (*medialive.ListInputSecurityGroupsOutput, error)
	ListInputSecurityGroupsRequest(*medialive.ListInputSecurityGroupsInput) (*request.Request, *medialive.ListInputSecurityGroupsOutput)

	ListInputSecurityGroupsPages(*medialive.ListInputSecurityGroupsInput, func(*medialive.ListInputSecurityGroupsOutput, bool) bool) error
	ListInputSecurityGroupsPagesWithContext(aws.Context, *medialive.ListInputSecurityGroupsInput, func(*medialive.ListInputSecurityGroupsOutput, bool) bool, ...request.Option) error

	ListInputs(*medialive.ListInputsInput) (*medialive.ListInputsOutput, error)
	ListInputsWithContext(aws.Context, *medialive.ListInputsInput, ...request.Option) (*medialive.ListInputsOutput, error)
	ListInputsRequest(*medialive.ListInputsInput) (*request.Request, *medialive.ListInputsOutput)

	ListInputsPages(*medialive.ListInputsInput, func(*medialive.ListInputsOutput, bool) bool) error
	ListInputsPagesWithContext(aws.Context, *medialive.ListInputsInput, func(*medialive.ListInputsOutput, bool) bool, ...request.Option) error

	StartChannel(*medialive.StartChannelInput) (*medialive.StartChannelOutput, error)
	StartChannelWithContext(aws.Context, *medialive.StartChannelInput, ...request.Option) (*medialive.StartChannelOutput, error)
	StartChannelRequest(*medialive.StartChannelInput) (*request.Request, *medialive.StartChannelOutput)

	StopChannel(*medialive.StopChannelInput) (*medialive.StopChannelOutput, error)
	StopChannelWithContext(aws.Context, *medialive.StopChannelInput, ...request.Option) (*medialive.StopChannelOutput, error)
	StopChannelRequest(*medialive.StopChannelInput) (*request.Request, *medialive.StopChannelOutput)

	UpdateChannel(*medialive.UpdateChannelInput) (*medialive.UpdateChannelOutput, error)
	UpdateChannelWithContext(aws.Context, *medialive.UpdateChannelInput, ...request.Option) (*medialive.UpdateChannelOutput, error)
	UpdateChannelRequest(*medialive.UpdateChannelInput) (*request.Request, *medialive.UpdateChannelOutput)
}

var _ MediaLiveAPI = (*medialive.MediaLive)(nil)
