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

	fmt.Printf("Image ready: %s\n", rawPath)
}
