package main

import (
	"context"
	"os"
	"path/filepath"

	"aws-tutorial/aws"
	"aws-tutorial/core/config"
	"aws-tutorial/core/logger"

	"go.uber.org/fx"
)

func main() {
	os.Setenv(aws.AWSConfigCredentialsEnv,
		filepath.Join(os.Getenv("HOME"), "projects", "aws", "credentials"))
	var (
		client aws.AWSClient
		log    logger.Logger
	)
	ctx := context.Background()
	app := fx.New(
		logger.Module,
		config.Module,
		aws.Module,
		fx.Populate(&client),
		fx.Populate(&log),
	)
	defer app.Stop(ctx) // nolint: errcheck

	if err := app.Start(ctx); err != nil {
		log.Fatal(err)
	}
	ss := client.Session()
	_, err := ss.Config.Credentials.Get()
	if err != nil {
		log.Fatal(err)
	}
	region := *ss.Config.Region
	log.Debugf("Region: %s", region)
}
