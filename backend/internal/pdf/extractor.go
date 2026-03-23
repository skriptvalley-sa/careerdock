// Package pdf extracts text content from PDF files.
//
// It uses pdfcpu for pure Go PDF text extraction. The extracted text is used
// as input for AI resume parsing and stored in the database.
// Claude receives the raw PDF bytes directly for ATS scoring.
package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// MaxFileSizeBytes is the maximum allowed PDF file size (5 MB).
const MaxFileSizeBytes = 5 * 1024 * 1024

// ExtractText extracts all text content from a PDF byte slice.
// Returns the concatenated text from all pages.
func ExtractText(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty PDF data")
	}

	// pdfcpu's ExtractContent writes files to a directory.
	// We use a temp dir and read back the output files.
	tmpDir, err := os.MkdirTemp("", "careerdock-pdf-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	r := bytes.NewReader(data)
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	if err := api.ExtractContent(r, tmpDir, "resume", nil, conf); err != nil {
		return "", fmt.Errorf("extract PDF content: %w", err)
	}

	// Read all extracted content files
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", fmt.Errorf("read temp dir: %w", err)
	}

	var parts []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
		if err != nil {
			continue
		}
		if text := strings.TrimSpace(string(content)); text != "" {
			parts = append(parts, text)
		}
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("PDF contains no extractable text (may be scanned/image-only)")
	}

	return normaliseWhitespace(strings.Join(parts, "\n\n")), nil
}

// IsPDF checks if the data starts with the PDF magic bytes (%PDF-).
func IsPDF(data []byte) bool {
	return len(data) >= 5 && string(data[:5]) == "%PDF-"
}

// normaliseWhitespace cleans up extracted text by collapsing excessive whitespace.
func normaliseWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	prevEmpty := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if !prevEmpty {
				result = append(result, "")
				prevEmpty = true
			}
			continue
		}
		prevEmpty = false
		result = append(result, trimmed)
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}
