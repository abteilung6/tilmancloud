package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func getEC2Client(ctx context.Context, region string) (*ec2.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := ec2.NewFromConfig(cfg)
	return client, nil
}

func main() {
	fmt.Println("Hello, World!")

	ctx := context.Background()

	ec2Client, err := getEC2Client(ctx, "eu-central-1")
	if err != nil {
		log.Fatalf("Failed to create EC2 client: %v", err)
	}

	fmt.Printf("AWS Config loaded successfully!\n")
	fmt.Printf("Region: eu-central-1\n")

	result, err := ec2Client.DescribeRegions(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to describe regions: %v", err)
	}

	fmt.Printf("Available regions (%d total)\n", len(result.Regions))

	fmt.Println("\n--- Creating EC2 Instance ---")

	runInput := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-004e960cde33f9146"),
		InstanceType: types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	runResult, err := ec2Client.RunInstances(ctx, runInput)
	if err != nil {
		log.Fatalf("Failed to run instance: %v", err)
	}

	if len(runResult.Instances) == 0 {
		log.Fatal("No instances were created")
	}
	instanceID := *runResult.Instances[0].InstanceId
	fmt.Printf("Instance launched! Instance ID: %s\n", instanceID)
	fmt.Printf("Current state: %s\n", runResult.Instances[0].State.Name)

	fmt.Println("Waiting for instance to be running...")
	maxWaitTime := 5 * time.Minute
	checkInterval := 10 * time.Second
	startTime := time.Now()

	for {
		if time.Since(startTime) > maxWaitTime {
			log.Fatal("Timeout waiting for instance to start")
		}

		describeInput := &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		}
		describeResult, err := ec2Client.DescribeInstances(ctx, describeInput)
		if err != nil {
			log.Fatalf("Failed to describe instance: %v", err)
		}

		if len(describeResult.Reservations) > 0 && len(describeResult.Reservations[0].Instances) > 0 {
			instance := describeResult.Reservations[0].Instances[0]
			state := instance.State.Name

			fmt.Printf("  Current state: %s\n", state)

			if state == types.InstanceStateNameRunning {
				fmt.Printf("✓ Instance is now running!\n")
				break
			}

			if state == types.InstanceStateNameTerminated || state == types.InstanceStateNameStopped {
				log.Fatalf("Instance entered %s state, failed to start", state)
			}
		}

		time.Sleep(checkInterval)
	}

	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	describeResult, err := ec2Client.DescribeInstances(ctx, describeInput)
	if err != nil {
		log.Fatalf("Failed to describe instance: %v", err)
	}

	if len(describeResult.Reservations) > 0 && len(describeResult.Reservations[0].Instances) > 0 {
		instance := describeResult.Reservations[0].Instances[0]
		fmt.Printf("\n--- Instance Details ---\n")
		fmt.Printf("Instance ID: %s\n", instanceID)
		fmt.Printf("Public IP: %s\n", getPtrStringValue(instance.PublicIpAddress))
		fmt.Printf("Private IP: %s\n", getPtrStringValue(instance.PrivateIpAddress))
		fmt.Printf("Instance Type: %s\n", instance.InstanceType)
		fmt.Printf("State: %s\n", instance.State.Name)
	}
}

func getPtrStringValue(ptr *string) string {
	if ptr == nil {
		return "N/A"
	}
	return *ptr
}
