#!/bin/bash -e

# You need to install the AWS Command Line Interface from http://aws.amazon.com/cli/
AMIID=$(aws ec2 describe-images --filters "Name=description, Values=Ubuntu 20.04 LTS" --query "Images[0].ImageId" --output text)
echo "Image ID: $AMIID"
VPCID=$(aws ec2 describe-vpcs --filter "Name=isDefault, Values=true" --query "Vpcs[0].VpcId" --output text)
echo "VPC ID: $VPCID"
SUBNETID=$(aws ec2 describe-subnets --filters "Name=vpc-id, Values=$VPCID" --query "Subnets[0].SubnetId" --output text)
echo "Subnet ID: $SUBNETID"
SGID=$(aws ec2 create-security-group --group-name awstutorial --description "Security group Ubuntu tutorial" --vpc-id $VPCID --output text)
echo "Security Group ID: $SGID"
aws ec2 authorize-security-group-ingress --group-id $SGID --protocol tcp --port 22 --cidr 0.0.0.0/0
INSTANCEID=$(aws ec2 run-instances --image-id $AMIID --key-name aws_tutorial --instance-type t2.micro --security-group-ids $SGID --subnet-id $SUBNETID --query "Instances[0].InstanceId" --output text)
echo "waiting for $INSTANCEID ..."
aws ec2 wait instance-running --instance-ids $INSTANCEID
PUBLICNAME=$(aws ec2 describe-instances --instance-ids $INSTANCEID --query "Reservations[0].Instances[0].PublicDnsName" --output text)
echo "$INSTANCEID is accepting SSH connections under $PUBLICNAME"
echo "ssh ubuntu@$PUBLICNAME"
read -p "Press [Enter] key to terminate $INSTANCEID ..."
aws ec2 terminate-instances --instance-ids $INSTANCEID
echo "terminating $INSTANCEID ..."
aws ec2 wait instance-terminated --instance-ids $INSTANCEID
aws ec2 delete-security-group --group-id $SGID
echo "done."
