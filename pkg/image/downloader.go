package image

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Downloader struct {
	buildDir string
	client   *http.Client
}

func NewDownloader(buildDir string) *Downloader {
	return &Downloader{
		buildDir: buildDir,
		client: &http.Client{
			Timeout: 30 * time.Minute, // Large files may take time
		},
	}
}

func (d *Downloader) Download(ctx context.Context, imageURL string) error {
	filename := deriveFilename(imageURL)
	compressedPath := filepath.Join(d.buildDir, filename)

	if exists, _ := d.IsDownloaded(filename); exists {
		slog.Info("Image already downloaded, skipping", "file", filename)
		return nil
	}

	if err := os.MkdirAll(d.buildDir, 0755); err != nil {
		return fmt.Errorf("create build directory: %w", err)
	}

	if err := d.downloadFile(ctx, imageURL, compressedPath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	slog.Info("Image downloaded successfully", "file", filename, "path", compressedPath)
	return nil
}

func (d *Downloader) downloadFile(ctx context.Context, url, destPath string) error {
	slog.Info("Downloading image", "url", url, "destination", destPath)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(destPath)
		return fmt.Errorf("write file: %w", err)
	}

	slog.Info("Download completed", "file", destPath, "size", written)
	return nil
}

func (d *Downloader) IsDownloaded(filename string) (bool, error) {
	path := filepath.Join(d.buildDir, filename)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.Size() > 0, nil
}

func deriveFilename(url string) string {
	base := filepath.Base(url)
	return base
}

func (d *Downloader) GetCompressedPath(imageURL string) string {
	filename := deriveFilename(imageURL)
	return filepath.Join(d.buildDir, filename)
}

func (d *Downloader) Decompress(ctx context.Context, compressedPath string) (string, error) {
	if !needsDecompression(compressedPath) {
		return "", fmt.Errorf("file does not need decompression (not a .xz file): %s", compressedPath)
	}

	rawPath := getRawPath(compressedPath)

	if exists, _ := d.isDecompressed(rawPath); exists {
		slog.Info("Image already decompressed, skipping", "file", rawPath)
		return rawPath, nil
	}

	if err := checkXZCommand(); err != nil {
		return "", fmt.Errorf("xz command not found: %w", err)
	}

	slog.Info("Decompressing image", "from", compressedPath, "to", rawPath)

	cmd := exec.CommandContext(ctx, "xz", "-d", "-c", compressedPath)

	out, err := os.Create(rawPath)
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()

	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Remove(rawPath)
		return "", fmt.Errorf("decompress failed: %w", err)
	}

	slog.Info("Decompression completed", "file", rawPath)
	return rawPath, nil
}

func needsDecompression(filename string) bool {
	return strings.HasSuffix(filename, ".xz")
}

func getRawPath(compressedPath string) string {
	return strings.TrimSuffix(compressedPath, ".xz")
}

func (d *Downloader) isDecompressed(rawPath string) (bool, error) {
	info, err := os.Stat(rawPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.Size() > 0, nil
}

func checkXZCommand() error {
	_, err := exec.LookPath("xz")
	if err != nil {
		return fmt.Errorf("xz command not found in PATH: %w", err)
	}
	return nil
}
