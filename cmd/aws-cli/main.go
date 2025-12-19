package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
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
}
