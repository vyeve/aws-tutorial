package aws

import (
	"errors"

	"aws-tutorial/core/config"
	"aws-tutorial/core/logger"

	"go.uber.org/fx"
)

var (
	ErrNoRunningInstances = errors.New("no running instances")
)

type Params struct {
	fx.In

	Logger logger.Logger
	Config *config.Configuration
}

type InstanceType int

const (
	Ubuntu20_04LTSx86 InstanceType = iota + 1
	Ubuntu20_04LTS_ARM
	Ubuntu18_04LTSx86
	Ubuntu18_04LTS_ARM
	AmazonLinux2x86
	AmazonLinux2ARM
	RedHat_x86
)

type ami struct {
	name string
	user string
	ami  string
}

// Free Tier images
var supportedInstanceTypes = map[InstanceType]ami{
	Ubuntu20_04LTSx86: {
		name: "Ubuntu Server 20.04 LTS x86",
		user: "ubuntu",
		ami:  "ami-089e6b3b328e5a2c1",
	},
	// Ubuntu20_04LTS_ARM: {
	// 	name: "Ubuntu Server 20.04 LTS ARM",
	// 	user: "ubuntu",
	// 	ami:  "ami-054e49cb26c2fd312",
	// },
	Ubuntu18_04LTSx86: {
		name: "Ubuntu Server 18.04 LTS x86",
		user: "ubuntu",
		ami:  "ami-00ddb0e5626798373",
	},
	// Ubuntu18_04LTS_ARM: {
	// 	name: "Ubuntu Server 18.04 LTS ARM",
	// 	user: "ubuntu",
	// 	ami:  "ami-074db80f0dc9b5f40",
	// },
	AmazonLinux2x86: {
		name: "Amazon Linux 2 AMI x86",
		user: "ec2-user",
		ami:  "ami-04d29b6f966df1537",
	},
	// AmazonLinux2ARM: {
	// 	name: "Amazon Linux 2 AMI ARM",
	// 	user: "ec2-user",
	// 	ami:  "ami-03156384f702d4eaf",
	// },
	RedHat_x86: {
		name: "Red Hat based image",
		user: "ec2-user",
		ami:  "ami-096fda3c22c1c990a",
	},
}

func (it InstanceType) IsValid() bool {
	_, ok := supportedInstanceTypes[it]
	return ok
}

func (it InstanceType) AMI() string {
	return supportedInstanceTypes[it].ami
}

func (it InstanceType) Name() string {
	return supportedInstanceTypes[it].name
}

func (it InstanceType) User() string {
	return supportedInstanceTypes[it].user
}
