package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func main() {
	fmt.Println("Hello, World!")

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-west-1"))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	fmt.Printf("AWS Config loaded successfully!\n")
	fmt.Printf("Region: %s\n", cfg.Region)

	ec2Client := ec2.NewFromConfig(cfg)

	result, err := ec2Client.DescribeRegions(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to describe regions: %v", err)
	}

	fmt.Printf("Available regions (%d total)\n", len(result.Regions))
}
