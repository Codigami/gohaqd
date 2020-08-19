// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

package marketplaceentitlementservice

import (
	"github.com/aws/aws-sdk-go/private/protocol"
)

const (

	// ErrCodeInternalServiceErrorException for service response error code
	// "InternalServiceErrorException".
	//
	// An internal error has occurred. Retry your request. If the problem persists,
	// post a message with details on the AWS forums.
	ErrCodeInternalServiceErrorException = "InternalServiceErrorException"

	// ErrCodeInvalidParameterException for service response error code
	// "InvalidParameterException".
	//
	// One or more parameters in your request was invalid.
	ErrCodeInvalidParameterException = "InvalidParameterException"

	// ErrCodeThrottlingException for service response error code
	// "ThrottlingException".
	//
	// The calls to the GetEntitlements API are throttled.
	ErrCodeThrottlingException = "ThrottlingException"
)

var exceptionFromCode = map[string]func(protocol.ResponseMetadata) error{
	"InternalServiceErrorException": newErrorInternalServiceErrorException,
	"InvalidParameterException":     newErrorInvalidParameterException,
	"ThrottlingException":           newErrorThrottlingException,
}
