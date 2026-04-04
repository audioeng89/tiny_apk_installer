package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

var tempADBDir string

func GetTempADBDir() string {
	if tempADBDir != "" {
		return tempADBDir
	}
	tempADBDir = filepath.Join(os.TempDir(), "tiny-apk-installer")
	return tempADBDir
}

func hasADBFiles(dir string, isWindows bool) error {
	exeName := "adb"
	if isWindows {
		exeName = "adb.exe"
	}

	if _, err := os.Stat(filepath.Join(dir, exeName)); os.IsNotExist(err) {
		return fmt.Errorf("adb not found")
	}

	if isWindows {
		if _, err := os.Stat(filepath.Join(dir, "AdbWinApi.dll")); os.IsNotExist(err) {
			return fmt.Errorf("AdbWinApi.dll not found")
		}
		if _, err := os.Stat(filepath.Join(dir, "AdbWinUsbApi.dll")); os.IsNotExist(err) {
			return fmt.Errorf("AdbWinUsbApi.dll not found")
		}
	}

	return nil
}

func ValidateADBPath(path string) error {
	return hasADBFiles(path, runtime.GOOS == "windows")
}

func CheckADBInTemp() (string, bool) {
	tempDir := GetTempADBDir()
	if err := hasADBFiles(tempDir, runtime.GOOS == "windows"); err == nil {
		exeName := "adb"
		if runtime.GOOS == "windows" {
			exeName = "adb.exe"
		}
		return filepath.Join(tempDir, exeName), true
	}
	return "", false
}

func DownloadADB() (string, error) {
	tempDir := GetTempADBDir()

	if err := os.RemoveAll(tempDir); err != nil {
		return "", fmt.Errorf("failed to clean temp directory: %w", err)
	}

	platform := runtime.GOOS
	var url string
	switch platform {
	case "windows":
		url = "https://dl.google.com/android/repository/platform-tools-latest-windows.zip"
	case "darwin":
		url = "https://dl.google.com/android/repository/platform-tools-latest-darwin.zip"
	case "linux":
		url = "https://dl.google.com/android/repository/platform-tools-latest-linux.zip"
	default:
		return "", fmt.Errorf("unsupported platform: %s", platform)
	}

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	zipPath := filepath.Join(os.TempDir(), "platform-tools.zip")

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	out, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write zip file: %w", err)
	}

	if err := extractZip(zipPath, tempDir); err != nil {
		return "", fmt.Errorf("failed to extract zip: %w", err)
	}
	os.Remove(zipPath)

	// move platform-tools contents to tempDir directly
	platformDir := filepath.Join(tempDir, "platform-tools")

	targetFiles := []string{"adb", "adb.exe", "AdbWinApi.dll", "AdbWinUsbApi.dll"}
	for _, f := range targetFiles {
		src := filepath.Join(platformDir, f)
		dst := filepath.Join(tempDir, f)
		if _, err := os.Stat(src); err == nil {
			os.Rename(src, dst)
			if platform != "windows" && f == "adb" {
				os.Chmod(dst, 0755)
			}
		}
	}

	os.RemoveAll(platformDir)

	exeName := "adb"
	if platform == "windows" {
		exeName = "adb.exe"
	}

	return filepath.Join(tempDir, exeName), nil
}
