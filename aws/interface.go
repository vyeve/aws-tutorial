package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type AWSClient interface {
	Session() *session.Session
	EC2(region string) *ec2.EC2
	CreateInstance(typ InstanceType, securityGroup string) (string, error)
	TerminateInstance(instanceID, securityGroup string) error
	Cleanup() error
}
