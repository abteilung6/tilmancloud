package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/abteilung6/tilmancloud/pkg/ec2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: command required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s <command>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands: create, list, delete\n")
		os.Exit(1)
	}

	command := os.Args[1]
	ctx := context.Background()

	ec2Client, err := ec2.NewClient(ctx, "eu-central-1")
	if err != nil {
		log.Fatalf("Failed to create EC2 client: %v", err)
	}

	switch command {
	case "create":
		instanceInfo, err := ec2.CreateInstance(ctx, ec2Client)
		if err != nil {
			log.Fatalf("Create command failed: %v", err)
		}

		if err := ec2.WaitForInstanceRunning(ctx, ec2Client, instanceInfo.InstanceID); err != nil {
			log.Fatalf("Failed to wait for instance: %v", err)
		}

		fmt.Printf("\n✓ Instance %s is now running!\n", instanceInfo.InstanceID)
	case "list":
		instances, err := ec2.ListInstances(ctx, ec2Client)
		if err != nil {
			log.Fatalf("List command failed: %v", err)
		}
		_ = instances
	case "delete":
		if len(os.Args) < 3 {
			log.Fatal("Delete command requires instance ID. Usage: delete <instance-id>")
		}
		instanceID := os.Args[2]
		if err := ec2.DeleteInstance(ctx, ec2Client, instanceID); err != nil {
			log.Fatalf("Delete command failed: %v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		os.Exit(1)
	}
}
