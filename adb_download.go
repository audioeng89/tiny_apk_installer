package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var tempADBDir string

func GetTempADBDir() string {
	if tempADBDir != "" {
		return tempADBDir
	}
	tempADBDir = filepath.Join(os.TempDir(), "tiny-apk-installer")
	return tempADBDir
}

func ValidateADBPath(path string) error {
	platform := runtime.GOOS
	exeName := getAdbExeName()

	adbPath := filepath.Join(path, exeName)
	if _, err := os.Stat(adbPath); os.IsNotExist(err) {
		return fmt.Errorf("adb not found in selected folder")
	}

	if platform == "windows" {
		dll1 := filepath.Join(path, "AdbWinApi.dll")
		dll2 := filepath.Join(path, "AdbWinUsbApi.dll")
		if _, err := os.Stat(dll1); os.IsNotExist(err) {
			return fmt.Errorf("AdbWinApi.dll not found in selected folder")
		}
		if _, err := os.Stat(dll2); os.IsNotExist(err) {
			return fmt.Errorf("AdbWinUsbApi.dll not found in selected folder")
		}
	}

	return nil
}

func CheckADBInTemp() (string, bool) {
	tempDir := GetTempADBDir()
	platform := runtime.GOOS
	exeName := getAdbExeName()

	adbPath := filepath.Join(tempDir, exeName)
	if _, err := os.Stat(adbPath); err == nil {
		if platform == "windows" {
			dll1 := filepath.Join(tempDir, "AdbWinApi.dll")
			dll2 := filepath.Join(tempDir, "AdbWinUsbApi.dll")
			if _, err := os.Stat(dll1); err == nil {
				if _, err := os.Stat(dll2); err == nil {
					return adbPath, true
				}
			}
		} else {
			return adbPath, true
		}
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

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer zr.Close()

	prefix := "platform-tools/"
	tempDir = filepath.Clean(tempDir)
	targetFiles := []string{"adb", "adb.exe", "AdbWinApi.dll", "AdbWinUsbApi.dll"}
	targetSet := make(map[string]bool)
	for _, f := range targetFiles {
		targetSet[f] = true
	}

	for _, f := range zr.File {
		if !strings.HasPrefix(f.Name, prefix) {
			continue
		}

		outName := strings.TrimPrefix(f.Name, prefix)
		if outName == "" {
			continue
		}

		baseName := outName
		if idx := strings.LastIndex(outName, "/"); idx > 0 {
			baseName = outName[idx+1:]
		}

		if !targetSet[baseName] {
			continue
		}

		destPath := filepath.Join(tempDir, outName)
		destPath = filepath.Clean(destPath)

		if !strings.HasPrefix(destPath, tempDir+string(filepath.Separator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		src, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open zip entry: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			src.Close()
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		dest, err := os.Create(destPath)
		if err != nil {
			src.Close()
			return "", fmt.Errorf("failed to create file: %w", err)
		}

		if _, err := io.Copy(dest, src); err != nil {
			src.Close()
			dest.Close()
			return "", fmt.Errorf("failed to extract file: %w", err)
		}

		src.Close()
		dest.Close()

		if platform != "windows" && (outName == "adb" || strings.HasSuffix(outName, "/adb")) {
			os.Chmod(destPath, 0755)
		}
	}

	os.Remove(zipPath)

	exeName := getAdbExeName()

	return filepath.Join(tempDir, exeName), nil
}
