package telegram

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// sanitizeFilename strips any path components and rejects dangerous filenames.
func sanitizeFilename(filename string) (string, error) {
	// Use only the base name — strip any directory components
	clean := filepath.Base(filename)

	// Reject empty, dot-only, or hidden files
	if clean == "" || clean == "." || clean == ".." {
		return "", fmt.Errorf("invalid filename: %q", filename)
	}

	// Reject if it still contains path separators or traversal
	if strings.ContainsAny(clean, `/\`) || strings.Contains(clean, "..") {
		return "", fmt.Errorf("invalid filename: %q", filename)
	}

	return clean, nil
}

func saveReceivedFile(dir, filename string, data []byte) (string, error) {
	clean, err := sanitizeFilename(filename)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create received dir: %w", err)
	}

	savePath := filepath.Join(dir, clean)

	// Verify the resolved path is still within the target directory
	absDir, _ := filepath.Abs(dir)
	absPath, _ := filepath.Abs(savePath)
	if !strings.HasPrefix(absPath, absDir+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal blocked: %q", filename)
	}

	// Avoid overwriting
	counter := 1
	for {
		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			break
		}
		ext := filepath.Ext(clean)
		stem := clean[:len(clean)-len(ext)]
		savePath = filepath.Join(dir, fmt.Sprintf("%s_%d%s", stem, counter, ext))
		counter++
	}

	if err := os.WriteFile(savePath, data, 0600); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return savePath, nil
}
