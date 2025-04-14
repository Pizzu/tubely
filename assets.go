package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type VideoInfoStreams struct {
	Streams []struct {
		Width  int `json:"width,omitempty"`
		Height int `json:"height,omitempty"`
	} `json:"streams"`
}

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(mediaType string) string {
	base := make([]byte, 32)
	_, err := rand.Read(base)
	if err != nil {
		panic("failed to generate random bytes")
	}
	id := base64.RawURLEncoding.EncodeToString(base)

	ext := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", id, ext)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	// Create a buffer to store the output for the command's standard output
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	err := cmd.Run()

	if err != nil {
		return "", err
	}

	var videoInfoStreams VideoInfoStreams

	err = json.Unmarshal(outBuf.Bytes(), &videoInfoStreams)

	if err != nil {
		return "", err
	}

	if len(videoInfoStreams.Streams) <= 0 {
		return "", errors.New("something went wrong getting video info")
	}

	width, height := videoInfoStreams.Streams[0].Width, videoInfoStreams.Streams[0].Height

	return getAspectRatio(width, height), nil
}

func getAspectRatio(width, height int) string {
	divisor := gcd(width, height)
	ratioWidth := width / divisor
	ratioHeight := height / divisor

	switch {
	case ratioWidth == 16 && ratioHeight == 9:
		return "landscape"
	case ratioWidth == 9 && ratioHeight == 16:
		return "portrait"
	default:
		return "other"
	}
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
