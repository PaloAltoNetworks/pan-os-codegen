package translate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loadTemplate loads a template from the templates/sdk directory
func loadTemplate(templatePath string) (string, error) {
	fullPath := filepath.Join("templates", "sdk", templatePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		// Try from parent directories (for when running from subdirectories)
		for i := 1; i <= 3; i++ {
			prefix := strings.Repeat("../", i)
			altPath := filepath.Join(prefix, "templates", "sdk", templatePath)
			content, err = os.ReadFile(altPath)
			if err == nil {
				break
			}
		}
		if err != nil {
			return "", fmt.Errorf("failed to read template %s: %w", fullPath, err)
		}
	}
	return string(content), nil
}

type structType string

const (
	structXmlType structType = "xml"
	structApiType structType = "api"
)
