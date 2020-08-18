// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

// Package secretsmanager provides the client and types for making API
// requests to AWS Secrets Manager.
//
// AWS Secrets Manager provides a service to enable you to store, manage, and
// retrieve, secrets.
//
// This guide provides descriptions of the Secrets Manager API. For more information
// about using this service, see the AWS Secrets Manager User Guide (https://docs.aws.amazon.com/secretsmanager/latest/userguide/introduction.html).
//
// API Version
//
// This version of the Secrets Manager API Reference documents the Secrets Manager
// API version 2017-10-17.
//
// As an alternative to using the API, you can use one of the AWS SDKs, which
// consist of libraries and sample code for various programming languages and
// platforms such as Java, Ruby, .NET, iOS, and Android. The SDKs provide a
// convenient way to create programmatic access to AWS Secrets Manager. For
// example, the SDKs provide cryptographically signing requests, managing errors,
// and retrying requests automatically. For more information about the AWS SDKs,
// including downloading and installing them, see Tools for Amazon Web Services
// (http://aws.amazon.com/tools/).
//
// We recommend you use the AWS SDKs to make programmatic API calls to Secrets
// Manager. However, you also can use the Secrets Manager HTTP Query API to
// make direct calls to the Secrets Manager web service. To learn more about
// the Secrets Manager HTTP Query API, see Making Query Requests (https://docs.aws.amazon.com/secretsmanager/latest/userguide/query-requests.html)
// in the AWS Secrets Manager User Guide.
//
// Secrets Manager API supports GET and POST requests for all actions, and doesn't
// require you to use GET for some actions and POST for others. However, GET
// requests are subject to the limitation size of a URL. Therefore, for operations
// that require larger sizes, use a POST request.
//
// Support and Feedback for AWS Secrets Manager
//
// We welcome your feedback. Send your comments to awssecretsmanager-feedback@amazon.com
// (mailto:awssecretsmanager-feedback@amazon.com), or post your feedback and
// questions in the AWS Secrets Manager Discussion Forum (http://forums.aws.amazon.com/forum.jspa?forumID=296).
// For more information about the AWS Discussion Forums, see Forums Help (http://forums.aws.amazon.com/help.jspa).
//
// How examples are presented
//
// The JSON that AWS Secrets Manager expects as your request parameters and
// the service returns as a response to HTTP query requests contain single,
// long strings without line breaks or white space formatting. The JSON shown
// in the examples displays the code formatted with both line breaks and white
// space to improve readability. When example input parameters can also cause
// long strings extending beyond the screen, you can insert line breaks to enhance
// readability. You should always submit the input as a single JSON text string.
//
// Logging API Requests
//
// AWS Secrets Manager supports AWS CloudTrail, a service that records AWS API
// calls for your AWS account and delivers log files to an Amazon S3 bucket.
// By using information that's collected by AWS CloudTrail, you can determine
// the requests successfully made to Secrets Manager, who made the request,
// when it was made, and so on. For more about AWS Secrets Manager and support
// for AWS CloudTrail, see Logging AWS Secrets Manager Events with AWS CloudTrail
// (http://docs.aws.amazon.com/secretsmanager/latest/userguide/monitoring.html#monitoring_cloudtrail)
// in the AWS Secrets Manager User Guide. To learn more about CloudTrail, including
// enabling it and find your log files, see the AWS CloudTrail User Guide (https://docs.aws.amazon.com/awscloudtrail/latest/userguide/what_is_cloud_trail_top_level.html).
//
// See https://docs.aws.amazon.com/goto/WebAPI/secretsmanager-2017-10-17 for more information on this service.
//
// See secretsmanager package documentation for more information.
// https://docs.aws.amazon.com/sdk-for-go/api/service/secretsmanager/
//
// Using the Client
//
// To contact AWS Secrets Manager with the SDK use the New function to create
// a new service client. With that client you can make API requests to the service.
// These clients are safe to use concurrently.
//
// See the SDK's documentation for more information on how to use the SDK.
// https://docs.aws.amazon.com/sdk-for-go/api/
//
// See aws.Config documentation for more information on configuring SDK clients.
// https://docs.aws.amazon.com/sdk-for-go/api/aws/#Config
//
// See the AWS Secrets Manager client SecretsManager for more
// information on creating client for this service.
// https://docs.aws.amazon.com/sdk-for-go/api/service/secretsmanager/#New
package secretsmanager
