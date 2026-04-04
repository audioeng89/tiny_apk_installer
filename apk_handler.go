package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type BundleType int

const (
	BundleNone BundleType = iota
	BundleAPKM
	BundleXAPK
	BundleAPKS
)

type APKHandler struct {
	bundleType  BundleType
	tempDir     string
	apkPaths    []string
	obbPath     string
	packageName string
	extracted   bool
}

type xapkManifest struct {
	PackageName string `json:"package_name"`
	VersionCode any    `json:"version_code"`
	Expansions  []struct {
		File string `json:"file"`
	} `json:"expansions"`
}

func DetectBundleType(path string) BundleType {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".apkm":
		return BundleAPKM
	case ".xapk":
		return BundleXAPK
	case ".apks":
		return BundleAPKS
	}
	return BundleNone
}

func ExtractBundle(path string, tempBaseDir string) (*APKHandler, error) {
	bundleType := DetectBundleType(path)
	if bundleType == BundleNone {
		return nil, fmt.Errorf("not a supported bundle format")
	}

	if err := os.MkdirAll(tempBaseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	handler := &APKHandler{
		bundleType: bundleType,
	}

	tempDir, err := os.MkdirTemp(tempBaseDir, "tiny-apk-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	handler.tempDir = tempDir

	if err := extractZip(path, tempDir); err != nil {
		handler.Cleanup()
		return nil, fmt.Errorf("failed to extract bundle: %w", err)
	}

	if err := handler.scanExtractedFiles(); err != nil {
		handler.Cleanup()
		return nil, err
	}

	handler.extracted = true
	return handler, nil
}

func (h *APKHandler) scanExtractedFiles() error {
	apkFiles, err := filepath.Glob(filepath.Join(h.tempDir, "*.apk"))
	if err != nil {
		return err
	}

	if len(apkFiles) == 0 {
		return fmt.Errorf("no APK files found in bundle")
	}

	h.apkPaths = apkFiles

	if h.bundleType == BundleXAPK {
		manifestPath := filepath.Join(h.tempDir, "manifest.json")
		if _, err := os.Stat(manifestPath); err == nil {
			if err := h.parseXAPKManifest(manifestPath); err != nil {
				return err
			}
		}

		obbBase := filepath.Join(h.tempDir, "Android", "obb")
		if _, err := os.Stat(obbBase); err == nil {
			h.obbPath = obbBase
		}
	}

	return nil
}

func (h *APKHandler) parseXAPKManifest(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var manifest xapkManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	h.packageName = manifest.PackageName
	return nil
}

func (h *APKHandler) Install(config ADBConfig) InstallResult {
	if !h.extracted {
		return InstallResult{Success: false, Error: "bundle not extracted"}
	}

	if len(h.apkPaths) == 0 {
		return InstallResult{Success: false, Error: "no APK files to install"}
	}

	result := config.InstallMultiple(h.apkPaths)
	if !result.Success {
		return result
	}

	if h.obbPath != "" && h.packageName != "" {
		pushResult := h.pushOBBFiles(config)
		if !pushResult.Success {
			return pushResult
		}
	}

	return InstallResult{Success: true}
}

func (h *APKHandler) pushOBBFiles(config ADBConfig) InstallResult {
	obbDest := filepath.Join("/sdcard/Android/obb", h.packageName)

	entries, err := os.ReadDir(h.obbPath)
	if err != nil {
		return InstallResult{Success: false, Error: "failed to read OBB directory"}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(h.obbPath, entry.Name())
		destPath := obbDest + "/" + entry.Name()

		_, err := runAdb("-s", config.Serial, "push", srcPath, destPath)
		if err != nil {
			return InstallResult{Success: false, Error: fmt.Sprintf("failed to push OBB: %v", err)}
		}
	}

	return InstallResult{Success: true}
}

func (h *APKHandler) Cleanup() {
	if h.tempDir != "" {
		os.RemoveAll(h.tempDir)
		h.tempDir = ""
	}
}

func (h *APKHandler) View() string {
	var b strings.Builder
	b.WriteString("APKs to Install:\n")
	for i, apk := range h.apkPaths {
		b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, filepath.Base(apk)))
	}
	if h.obbPath != "" {
		b.WriteString(fmt.Sprintf("\nOBB Data: %s\n", filepath.Base(h.obbPath)))
	}
	return b.String()
}

type BundleExtractedMsg struct {
	Handler *APKHandler
	Err     error
}
