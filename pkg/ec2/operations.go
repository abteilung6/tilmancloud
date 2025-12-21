package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func WaitForInstanceRunning(ctx context.Context, client EC2Client, instanceID string) error {
	fmt.Println("Waiting for instance to be running...")
	maxWaitTime := 5 * time.Minute
	checkInterval := 10 * time.Second
	startTime := time.Now()

	for {
		if time.Since(startTime) > maxWaitTime {
			return fmt.Errorf("timeout waiting for instance to start")
		}

		describeInput := &awsec2.DescribeInstancesInput{
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

func CreateInstance(ctx context.Context, client EC2Client) (string, error) {
	fmt.Println("--- Creating EC2 Instance ---")

	runInput := &awsec2.RunInstancesInput{
		ImageId:      aws.String("ami-004e960cde33f9146"),
		InstanceType: types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	runResult, err := client.RunInstances(ctx, runInput)
	if err != nil {
		return "", fmt.Errorf("failed to run instance: %w", err)
	}

	if len(runResult.Instances) == 0 {
		return "", fmt.Errorf("no instances were created")
	}

	instanceID := *runResult.Instances[0].InstanceId
	fmt.Printf("Instance launched! Instance ID: %s\n", instanceID)
	fmt.Printf("Current state: %s\n", runResult.Instances[0].State.Name)

	return instanceID, nil
}

func ListInstances(ctx context.Context, client EC2Client) error {
	fmt.Println("--- Listing EC2 Instances ---")

	describeInput := &awsec2.DescribeInstancesInput{}
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

func DeleteInstance(ctx context.Context, client EC2Client, instanceID string) error {
	fmt.Printf("--- Deleting EC2 Instance: %s ---\n", instanceID)

	terminateInput := &awsec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	terminateResult, err := client.TerminateInstances(ctx, terminateInput)
	if err != nil {
		return fmt.Errorf("failed to terminate instance: %w", err)
	}

	if len(terminateResult.TerminatingInstances) == 0 {
		return fmt.Errorf("no instances were terminated")
	}

	instanceState := terminateResult.TerminatingInstances[0]
	fmt.Printf("Termination initiated for instance: %s\n", *instanceState.InstanceId)
	fmt.Printf("Current state: %s -> %s\n", instanceState.PreviousState.Name, instanceState.CurrentState.Name)
	fmt.Println("\nInstance termination in progress...")
	return nil
}

func getPtrStringValue(ptr *string) string {
	if ptr == nil {
		return "N/A"
	}
	return *ptr
}
