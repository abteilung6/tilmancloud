package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/abteilung6/tilmancloud/pkg/image"
)

func main() {
	runBuild()
}

func runBuild() {
	ctx := context.Background()
	downloader := image.NewDownloader("build/images")
	cloud_base_image_url := "https://download.fedoraproject.org/pub/fedora/linux/releases/43/Server/aarch64/images/Fedora-Server-Host-Generic-43-1.6.aarch64.raw.xz"

	err := downloader.Download(ctx, cloud_base_image_url)
	if err != nil {
		slog.Error("Failed to download image", "error", err)
		os.Exit(1)
	}

	compressedPath := downloader.GetCompressedPath(cloud_base_image_url)
	rawPath, err := downloader.Decompress(ctx, compressedPath)
	if err != nil {
		slog.Error("Failed to decompress image", "error", err)
		os.Exit(1)
	}

	imageID, err := image.GenerateImageIDFromFile(rawPath)
	if err != nil {
		slog.Error("Failed to generate ImageID", "error", err)
		os.Exit(1)
	}
	slog.Info("Generated ImageID", "image_id", imageID)

	bucket := os.Getenv("AWS_S3_BUCKET")
	if bucket == "" {
		slog.Error("AWS_S3_BUCKET environment variable not set")
		os.Exit(1)
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "eu-central-1"
	}

	uploader, err := image.NewS3Uploader(ctx, bucket, region)
	if err != nil {
		slog.Error("Failed to create S3 uploader", "error", err)
		os.Exit(1)
	}

	s3Key := image.GenerateS3Key(rawPath)
	err = uploader.Upload(ctx, rawPath, s3Key)
	if err != nil {
		slog.Error("Failed to upload image to S3", "error", err)
		os.Exit(1)
	}
	fmt.Printf("Image uploaded to S3: %s\n", uploader.GetS3URL(s3Key))

	importer, err := image.NewImporter(ctx, region)
	if err != nil {
		slog.Error("Failed to create importer", "error", err)
		os.Exit(1)
	}

	description := "Fedora 43 aarch64 base image"
	snapshotID, err := importer.ImportSnapshot(ctx, bucket, s3Key, description, imageID)
	if err != nil {
		slog.Error("Failed to import snapshot", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Snapshot created: %s\n", snapshotID)

	registrar, err := image.NewAMIRegistrar(ctx, region)
	if err != nil {
		slog.Error("Failed to create AMI registrar", "error", err)
		os.Exit(1)
	}

	amiName := fmt.Sprintf("fedora-43-aarch64-base-%s", imageID)
	amiID, err := registrar.RegisterAMI(ctx, snapshotID, imageID, amiName, description)
	if err != nil {
		slog.Error("Failed to register AMI", "error", err)
		os.Exit(1)
	}

	fmt.Printf("AMI registered: %s\n", amiID)

}
