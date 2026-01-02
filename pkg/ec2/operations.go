package ec2

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type InstanceInfo struct {
	InstanceID   string
	State        string
	InstanceType string
	PublicIP     string
	PrivateIP    string
}

func WaitForInstanceRunning(ctx context.Context, client EC2Client, instanceID string) error {
	slog.Info("Waiting for instance to be running", "instance_id", instanceID)
	maxWaitTime := 5 * time.Minute
	checkInterval := 10 * time.Second
	startTime := time.Now()

	for {
		if time.Since(startTime) > maxWaitTime {
			slog.Error("Timeout waiting for instance to start", "instance_id", instanceID, "timeout", maxWaitTime)
			return fmt.Errorf("timeout waiting for instance to start")
		}

		describeInput := &awsec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		}
		describeResult, err := client.DescribeInstances(ctx, describeInput)
		if err != nil {
			slog.Error("Failed to describe instance", "instance_id", instanceID, "error", err)
			return fmt.Errorf("failed to describe instance: %w", err)
		}

		if len(describeResult.Reservations) > 0 && len(describeResult.Reservations[0].Instances) > 0 {
			instance := describeResult.Reservations[0].Instances[0]
			state := instance.State.Name

			slog.Debug("Instance state check", "instance_id", instanceID, "state", state)

			if state == types.InstanceStateNameRunning {
				slog.Info("Instance is now running", "instance_id", instanceID)
				return nil
			}

			if state == types.InstanceStateNameTerminated || state == types.InstanceStateNameStopped {
				slog.Error("Instance entered invalid state", "instance_id", instanceID, "state", state)
				return fmt.Errorf("instance entered %s state, failed to start", state)
			}
		}

		time.Sleep(checkInterval)
	}
}

type CreateInstanceConfig struct {
	ImageID      string
	InstanceType types.InstanceType
}

func CreateInstance(ctx context.Context, client EC2Client, config CreateInstanceConfig) (InstanceInfo, error) {
	slog.Info("Creating EC2 instance", "image_id", config.ImageID, "instance_type", config.InstanceType)

	runInput := &awsec2.RunInstancesInput{
		ImageId:      aws.String(config.ImageID),
		InstanceType: config.InstanceType,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	runResult, err := client.RunInstances(ctx, runInput)
	if err != nil {
		slog.Error("Failed to run instance", "error", err)
		return InstanceInfo{}, fmt.Errorf("failed to run instance: %w", err)
	}

	if len(runResult.Instances) == 0 {
		slog.Error("No instances were created")
		return InstanceInfo{}, fmt.Errorf("no instances were created")
	}

	instance := runResult.Instances[0]
	info := InstanceInfo{
		InstanceID:   getPtrStringValue(instance.InstanceId),
		State:        string(instance.State.Name),
		InstanceType: string(instance.InstanceType),
		PublicIP:     getPtrStringValue(instance.PublicIpAddress),
		PrivateIP:    getPtrStringValue(instance.PrivateIpAddress),
	}

	slog.Info("Instance created successfully",
		"instance_id", info.InstanceID,
		"state", info.State,
		"instance_type", info.InstanceType)

	return info, nil
}

func ListInstances(ctx context.Context, client EC2Client) ([]InstanceInfo, error) {
	slog.Debug("Listing EC2 instances")

	describeInput := &awsec2.DescribeInstancesInput{}
	describeResult, err := client.DescribeInstances(ctx, describeInput)
	if err != nil {
		slog.Error("Failed to describe instances", "error", err)
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	var instances []InstanceInfo
	for _, reservation := range describeResult.Reservations {
		for _, instance := range reservation.Instances {
			info := InstanceInfo{
				InstanceID:   getPtrStringValue(instance.InstanceId),
				State:        string(instance.State.Name),
				InstanceType: string(instance.InstanceType),
				PublicIP:     getPtrStringValue(instance.PublicIpAddress),
				PrivateIP:    getPtrStringValue(instance.PrivateIpAddress),
			}
			instances = append(instances, info)
		}
	}

	slog.Info("Listed instances", "count", len(instances))
	return instances, nil
}

func DeleteInstance(ctx context.Context, client EC2Client, instanceID string) error {
	slog.Info("Deleting EC2 instance", "instance_id", instanceID)

	terminateInput := &awsec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	terminateResult, err := client.TerminateInstances(ctx, terminateInput)
	if err != nil {
		slog.Error("Failed to terminate instance", "instance_id", instanceID, "error", err)
		return fmt.Errorf("failed to terminate instance: %w", err)
	}

	if len(terminateResult.TerminatingInstances) == 0 {
		slog.Error("No instances were terminated", "instance_id", instanceID)
		return fmt.Errorf("no instances were terminated")
	}

	instanceState := terminateResult.TerminatingInstances[0]
	slog.Info("Instance termination initiated",
		"instance_id", *instanceState.InstanceId,
		"previous_state", instanceState.PreviousState.Name,
		"current_state", instanceState.CurrentState.Name)

	return nil
}

func getPtrStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
