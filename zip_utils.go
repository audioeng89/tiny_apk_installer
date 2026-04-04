package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func extractZip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	destDir = filepath.Clean(destDir)

	for _, file := range reader.File {
		path := filepath.Join(destDir, file.Name)
		path = filepath.Clean(path)

		if !strings.HasPrefix(path, destDir+string(filepath.Separator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			return err
		}

		writer, err := os.Create(path)
		if err != nil {
			rc.Close()
			return err
		}

		if _, err := io.Copy(writer, rc); err != nil {
			rc.Close()
			writer.Close()
			return err
		}

		rc.Close()
		writer.Close()
	}

	return nil
}
