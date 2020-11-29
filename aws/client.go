package aws

import (
	"aws-tutorial/core/config"
	"aws-tutorial/core/logger"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

const AWSConfigCredentialsEnv = "AWS_CONFIG_CREDENTIALS"

type awsClient struct {
	ses    *session.Session
	logger logger.Logger
	conf   *config.Configuration
}

func New(param Params) (AWSClient, error) {
	c := &awsClient{
		logger: param.Logger,
		conf:   param.Config,
	}
	var err error
	c.ses, err = session.NewSession(&aws.Config{
		Region:                        aws.String(c.conf.Region),
		Credentials:                   credentials.NewSharedCredentials(os.Getenv(AWSConfigCredentialsEnv), c.conf.Profile),
		CredentialsChainVerboseErrors: &c.conf.CredentialsChainVerboseErrors,
	})
	if err != nil {
		c.logger.Errorf("Failed to create new session. err: %v", err)
		return nil, err
	}
	c.logger.Info("New AWS session created...")
	c.logger.Debugf("Config: %+v", param.Config)
	return c, nil
}

func (c *awsClient) Session() *session.Session {
	return c.ses
}
