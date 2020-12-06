package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	create(client, log)
	wait()
	fullCleanup(client, log)
}

func wait() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter KILL to exit")
	for {
		text, _ := reader.ReadString('\n')
		text = strings.ToLower(strings.TrimSpace(text))
		switch text {
		case "q", "quit", "stop", "kill", "exit":
			return
		default:
			fmt.Println("Enter KILL to exit")
		}

	}
}

func create(client aws.AWSClient, log logger.Logger) {
	sg := "mySG-001"
	instanceID, err := client.CreateInstance(aws.Ubuntu20_04LTSx86, sg)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Created instance %s", instanceID)

}

func fullCleanup(client aws.AWSClient, log logger.Logger) {
	err := client.Cleanup()
	if err != nil {
		log.Fatal(err)
	}
}
