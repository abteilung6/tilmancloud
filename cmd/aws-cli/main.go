package main

import (
	"context"
	"fmt"
	"log"
	"os"
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

func waitForInstanceRunning(ctx context.Context, client *ec2.Client, instanceID string) error {
	fmt.Println("Waiting for instance to be running...")
	maxWaitTime := 5 * time.Minute
	checkInterval := 10 * time.Second
	startTime := time.Now()

	for {
		if time.Since(startTime) > maxWaitTime {
			return fmt.Errorf("timeout waiting for instance to start")
		}

		describeInput := &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		}
		describeResult, err := client.DescribeInstances(ctx, describeInput)
		if err != nil {
			return fmt.Errorf("failed to describe instance: %w", err)
		}

		if len(describeResult.Reservations) > 0 && len(describeResult.Reservations[0].Instances) > 0 {
			instance := describeResult.Reservations[0].Instances[0]
			state := instance.State.Name

			fmt.Printf("  Current state: %s\n", state)

			if state == types.InstanceStateNameRunning {
				fmt.Printf("✓ Instance is now running!\n")
				return nil
			}

			if state == types.InstanceStateNameTerminated || state == types.InstanceStateNameStopped {
				return fmt.Errorf("instance entered %s state, failed to start", state)
			}
		}

		time.Sleep(checkInterval)
	}
}

func cmdCreate(ctx context.Context, client *ec2.Client) error {
	fmt.Println("--- Creating EC2 Instance ---")

	runInput := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-004e960cde33f9146"),
		InstanceType: types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	runResult, err := client.RunInstances(ctx, runInput)
	if err != nil {
		return fmt.Errorf("failed to run instance: %w", err)
	}

	if len(runResult.Instances) == 0 {
		return fmt.Errorf("no instances were created")
	}

	instanceID := *runResult.Instances[0].InstanceId
	fmt.Printf("Instance launched! Instance ID: %s\n", instanceID)
	fmt.Printf("Current state: %s\n", runResult.Instances[0].State.Name)

	if err := waitForInstanceRunning(ctx, client, instanceID); err != nil {
		return err
	}

	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	describeResult, err := client.DescribeInstances(ctx, describeInput)
	if err != nil {
		return fmt.Errorf("failed to describe instance: %w", err)
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

	return nil
}

func cmdList(ctx context.Context, client *ec2.Client) error {
	fmt.Println("--- Listing EC2 Instances ---")

	describeInput := &ec2.DescribeInstancesInput{}
	describeResult, err := client.DescribeInstances(ctx, describeInput)
	if err != nil {
		return fmt.Errorf("failed to describe instances: %w", err)
	}

	totalInstances := 0
	for _, reservation := range describeResult.Reservations {
		totalInstances += len(reservation.Instances)
	}

	if totalInstances == 0 {
		fmt.Println("No instances found.")
		return nil
	}

	fmt.Printf("\nFound %d instance(s):\n\n", totalInstances)
	fmt.Printf("%-20s %-15s %-18s %-18s %-12s\n", "Instance ID", "State", "Type", "Public IP", "Private IP")
	fmt.Println("--------------------------------------------------------------------------------")

	for _, reservation := range describeResult.Reservations {
		for _, instance := range reservation.Instances {
			instanceID := getPtrStringValue(instance.InstanceId)
			state := string(instance.State.Name)
			instanceType := string(instance.InstanceType)
			publicIP := getPtrStringValue(instance.PublicIpAddress)
			privateIP := getPtrStringValue(instance.PrivateIpAddress)

			fmt.Printf("%-20s %-15s %-18s %-18s %-12s\n",
				instanceID, state, instanceType, publicIP, privateIP)
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: command required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s <command>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands: create, list\n")
		os.Exit(1)
	}

	command := os.Args[1]
	ctx := context.Background()

	ec2Client, err := getEC2Client(ctx, "eu-central-1")
	if err != nil {
		log.Fatalf("Failed to create EC2 client: %v", err)
	}

	switch command {
	case "create":
		if err := cmdCreate(ctx, ec2Client); err != nil {
			log.Fatalf("Create command failed: %v", err)
		}
	case "list":
		if err := cmdList(ctx, ec2Client); err != nil {
			log.Fatalf("List command failed: %v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		os.Exit(1)
	}
}

func getPtrStringValue(ptr *string) string {
	if ptr == nil {
		return "N/A"
	}
	return *ptr
}
