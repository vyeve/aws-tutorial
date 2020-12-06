package aws

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"aws-tutorial/core/config"
	"aws-tutorial/core/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const AWSConfigCredentialsEnv = "AWS_CONFIG_CREDENTIALS"

type awsClient struct {
	ses    *session.Session
	logger logger.Logger
	conf   *config.Configuration
	ec2    *ec2.EC2
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
	c.ec2 = ec2.New(c.ses)
	return c, nil
}

func (c *awsClient) Session() *session.Session {
	return c.ses
}

func (c *awsClient) EC2(region string) *ec2.EC2 {
	var configs []*aws.Config
	if len(region) != 0 {
		configs = append(configs, aws.NewConfig().WithRegion(region))
	}
	return ec2.New(c.ses, configs...)
}

func (c *awsClient) CreateInstance(typ InstanceType, securityGroup string) (string, error) {
	if !typ.IsValid() {
		return "", fmt.Errorf("unsupported instance type: %d", typ)
	}
	ami, err := c.validateImage(typ)
	if err != nil {
		return "", err
	}
	c.logger.Debugf("found AMI: %s for %s", ami, typ.Name())
	vpc, subnet, err := c.getDefaultVPC()
	if err != nil {
		return "", err
	}
	c.logger.Debugf("found default VPC: %s. SubnetID: %s", vpc, subnet)
	sgID, err := c.createSecurityGroup(securityGroup, vpc)
	if err != nil {
		return "", err
	}
	c.logger.Infof("created Security Group for VPC: %s. SecurityGroupID: %s", vpc, sgID)
	instanceID, err := c.createInstance(ami, sgID, subnet)
	if err != nil {
		return "", err
	}
	c.logger.Infof("created instance %s. InstanceID: %s", typ.Name(), instanceID)
	instance, err := c.getInstance(instanceID)
	if err != nil {
		return "", err
	}
	c.logger.Infof("connect: ssh %s@%s", typ.User(), *instance.PublicIpAddress)
	return instanceID, nil
}

func (c *awsClient) TerminateInstance(instanceID, securityGroup string) error {
	err := c.terminateInstances(instanceID)
	if err != nil {
		return err
	}
	return c.deleteGroup(securityGroup)
}

func (c *awsClient) Cleanup() error {
	instances, err := c.listRunningInstances()
	switch err {
	case ErrNoRunningInstances:
	case nil:
		instanceIDs := make([]string, 0, len(instances))
		for _, i := range instances {
			instanceIDs = append(instanceIDs, *i.InstanceId)
			c.logger.Debugf("InstanceID: %s, State: %s", *i.InstanceId, *i.State.Name)
		}

		err = c.terminateInstances(instanceIDs...)
		if err != nil {
			return err
		}
	default:
		return err
	}
	err = c.deleteSecurityGroups()
	if err != nil {
		return err
	}
	return c.deleteVolumes()
}

func (c *awsClient) validateImage(typ InstanceType) (string, error) {
	images, err := c.ec2.DescribeImages(&ec2.DescribeImagesInput{
		DryRun:   aws.Bool(false),
		ImageIds: aws.StringSlice([]string{typ.AMI()}),
	})
	if err != nil {
		return "", err
	}
	if len(images.Images) != 1 {
		return "", fmt.Errorf("unexpected result for %s AMI: %d", typ.AMI(), len(images.Images))
	}
	if *images.Images[0].ImageId != typ.AMI() {
		return "", fmt.Errorf("unexpected result: %s != %s", typ.AMI(), *images.Images[0].ImageId)
	}
	return typ.AMI(), nil
}

func (c *awsClient) getDefaultVPC() (string, string, error) {
	vpcs, err := c.ec2.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: aws.StringSlice([]string{"true"}),
			},
		},
	})
	if err != nil {
		return "", "", err
	}
	if len(vpcs.Vpcs) == 0 {
		return "", "", fmt.Errorf("cannot find any VPCS")
	}
	vpcID := *vpcs.Vpcs[0].VpcId
	subnet, err := c.ec2.DescribeSubnets(&ec2.DescribeSubnetsInput{
		DryRun: aws.Bool(false),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: aws.StringSlice([]string{vpcID}),
			},
		},
	})
	if err != nil {
		return "", "", err
	}
	if len(vpcs.Vpcs) == 0 {
		return "", "", fmt.Errorf("cannot find any subnet for VPC: %s", vpcID)
	}
	for _, sn := range subnet.Subnets {
		c.logger.Debugf("Subnet: %s, Zone: %s", *sn.SubnetId, *sn.AvailabilityZone)

	}
	return vpcID, *subnet.Subnets[rand.Intn(len(subnet.Subnets))].SubnetId, nil
}

func (c *awsClient) createSecurityGroup(groupName, vpcsID string) (string, error) {
	out, err := c.ec2.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		Description: aws.String(fmt.Sprintf("New Security group: %s", groupName)),
		TagSpecifications: []*ec2.TagSpecification{
			{
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("aws"),
						Value: aws.String("golang_SDK"),
					},
				},
				ResourceType: aws.String("security-group"),
			},
		},
		VpcId:     &vpcsID,
		GroupName: aws.String(groupName),
	})
	if err != nil {
		return "", err
	}
	_, err = c.ec2.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    out.GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(22),
		ToPort:     aws.Int64(22),
		CidrIp:     aws.String("0.0.0.0/0"),
	})
	return *out.GroupId, err
}

func (c *awsClient) createInstance(ami, sgID, subnetID string) (string, error) {
	out, err := c.ec2.RunInstances(&ec2.RunInstancesInput{
		ImageId:          &ami,
		KeyName:          aws.String("aws_tutorial"),
		InstanceType:     aws.String("t2.micro"),
		SecurityGroupIds: aws.StringSlice([]string{sgID}),
		SubnetId:         &subnetID,
		MaxCount:         aws.Int64(1),
		MinCount:         aws.Int64(1),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &ec2.EbsBlockDevice{
					DeleteOnTermination: aws.Bool(true),
					Encrypted:           aws.Bool(false),
					VolumeSize:          aws.Int64(9),
					VolumeType:          aws.String("gp2"),
				},
			},
		},
	})
	if err != nil {
		return "", err
	}
	if len(out.Instances) != 1 {
		return "", fmt.Errorf("unexpected result: %d", len(out.Instances))
	}
	instanceID := *out.Instances[0].InstanceId
	c.logger.Debugf("Wait until run for instance %s...", instanceID)
	stop := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go wait(&wg, stop)
	err = c.ec2.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{instanceID}),
	})
	if err != nil {
		return "", err
	}
	close(stop)
	wg.Wait()
	return instanceID, nil
}

func (c *awsClient) getInstance(instanceID string) (*ec2.Instance, error) {
	out, err := c.ec2.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{instanceID}),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Reservations) != 1 || len(out.Reservations[0].Instances) != 1 {
		return nil, fmt.Errorf("unexpected result for instance %s", instanceID)
	}
	return out.Reservations[0].Instances[0], nil
}

func (c *awsClient) terminateInstances(instanceIDs ...string) error {
	c.logger.Debugf("Try to terminate instance %s ...", instanceIDs)
	_, err := c.ec2.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice(instanceIDs),
	})
	if err != nil {
		return err
	}
	stop := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go wait(&wg, stop)
	err = c.ec2.WaitUntilInstanceTerminated(&ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice(instanceIDs),
	})
	close(stop)
	wg.Wait()
	return err
}

func (c *awsClient) deleteGroup(securityGroup string) error {
	_, err := c.ec2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
		GroupName: &securityGroup,
	})
	return err
}

func wait(wg *sync.WaitGroup, stop <-chan struct{}) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			fmt.Printf(" .")
		case <-stop:
			fmt.Println()
			return
		}
	}
}

func (c *awsClient) listRunningInstances() ([]*ec2.Instance, error) {
	// https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-instances.html
	// 0 : pending
	// 16 : running
	// 32 : shutting-down
	// 48 : terminated
	// 64 : stopping
	// 80 : stopped

	res, err := c.ec2.DescribeInstances(&ec2.DescribeInstancesInput{
		MaxResults: aws.Int64(1000),
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-state-code"),
				Values: aws.StringSlice([]string{
					"0",  /*pending*/
					"16", /*running*/
					"32", /*shutting-down*/
					// "48", /*terminated*/
					"64", /*stopping*/
					"80", /*stopped*/
				}),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	instances := make([]*ec2.Instance, 0)
	for _, reservation := range res.Reservations {
		instances = append(instances, reservation.Instances...)
	}
	if len(instances) == 0 {
		return nil, ErrNoRunningInstances
	}
	return instances, nil
}

func (c *awsClient) deleteSecurityGroups() error {
	res, err := c.ec2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		c.logger.Warn(err)
		return err
	}
	groups := make([]string, 0, len(res.SecurityGroups))
	for _, sg := range res.SecurityGroups {
		c.logger.Debugf("SecurityGroup: %s, %s", *sg.GroupId, *sg.GroupName)
		if *sg.GroupName == "default" {
			continue
		}
		_, err = c.ec2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
			GroupId: sg.GroupId,
		})
		if err != nil {
			return err
		}
		groups = append(groups, *sg.GroupId)
	}
	return nil
}

func (c *awsClient) deleteVolumes() error {
	res, err := c.ec2.DescribeVolumes(&ec2.DescribeVolumesInput{
		MaxResults: aws.Int64(1000),
	})
	if err != nil {
		return err
	}
	for _, vol := range res.Volumes {
		c.logger.Debugf("VolumeID: %s, Size: %dGiB", *vol.VolumeId, *vol.Size)
		_, err = c.ec2.DeleteVolume(&ec2.DeleteVolumeInput{
			VolumeId: vol.VolumeId,
		})
		if err != nil {
			return err
		}
	}
	return err
}
