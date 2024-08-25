package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	AVAILABILITYZONE = "ap-southeast-1b"
	VPCNAME          = "my-vpc-asad"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a new VPC
		vpc, err := ec2.NewVpc(ctx, VPCNAME, &ec2.VpcArgs{
			CidrBlock: pulumi.String("10.0.0.0/16"), // CIDR block for the VPC
			Tags: pulumi.StringMap{
				"Name": pulumi.String(VPCNAME),
			},
			EnableDnsHostnames: pulumi.Bool(true),
			EnableDnsSupport:   pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// Create a public subnet
		publicSubnet, err := ec2.NewSubnet(ctx, "public-subnet", &ec2.SubnetArgs{
			VpcId:               vpc.ID(),
			CidrBlock:           pulumi.String("10.0.1.0/24"),    // CIDR block for the public subnet
			MapPublicIpOnLaunch: pulumi.Bool(true),               // Assign public IP addresses to instances
			AvailabilityZone:    pulumi.String(AVAILABILITYZONE), // Specify the availability zone
		})
		if err != nil {
			return err
		}

		// Create a private subnet
		privateSubnet, err := ec2.NewSubnet(ctx, "private-subnet", &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String("10.0.2.0/24"),    // CIDR block for the private subnet
			AvailabilityZone: pulumi.String(AVAILABILITYZONE), // Specify the availability zone
		})
		if err != nil {
			return err
		}

		// Create an Internet Gateway for the VPC
		igw, err := ec2.NewInternetGateway(ctx, "internet-gateway", &ec2.InternetGatewayArgs{
			VpcId: vpc.ID(),
		})
		if err != nil {
			return err
		}

		// Create a route table for the public subnet
		publicRouteTable, err := ec2.NewRouteTable(ctx, "public-route-table", &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
			Routes: ec2.RouteTableRouteArray{
				ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"), // Allow outbound traffic to the internet
					GatewayId: igw.ID(),                   // Specify the Internet Gateway
				},
			},
		})
		if err != nil {
			return err
		}

		// Associate the public subnet with the route table
		_, err = ec2.NewRouteTableAssociation(ctx, "public-route-table-association", &ec2.RouteTableAssociationArgs{
			SubnetId:     publicSubnet.ID(),
			RouteTableId: publicRouteTable.ID(),
		})
		if err != nil {
			return err
		}

		// Create a security group that allows SSH access
		sg, err := ec2.NewSecurityGroup(ctx, "security-group", &ec2.SecurityGroupArgs{
			VpcId:       vpc.ID(),
			Description: pulumi.String("Allow SSH"),
			Ingress: ec2.SecurityGroupIngressArray{
				ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(22),
					ToPort:     pulumi.Int(22),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")}, // Allow SSH from any IP address
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				ec2.SecurityGroupEgressArgs{
					Protocol:   pulumi.String("-1"),
					FromPort:   pulumi.Int(0),
					ToPort:     pulumi.Int(0),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")}, // Allow all outbound traffic
				},
			},
		})
		if err != nil {
			return err
		}

		// Export the VPC and Subnet IDs
		ctx.Export("vpcId", vpc.ID())
		ctx.Export("publicSubnetId", publicSubnet.ID())
		ctx.Export("privateSubnetId", privateSubnet.ID())
		ctx.Export("securityGroupId", sg.ID())

		return nil
	})
}
