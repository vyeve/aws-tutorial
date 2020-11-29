package aws

import "github.com/aws/aws-sdk-go/aws/session"

type AWSClient interface {
	Session() *session.Session
}
