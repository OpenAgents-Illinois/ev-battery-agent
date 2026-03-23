package agent

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// textChunk is a piece of text extracted from a document.
type textChunk struct {
	text string
}

var batteryKeywords = []string{
	"battery", "thermal", "voltage", "charging", "temperature", "cell",
	"overheating", "capacity", "kilowatt", "kwh", "high voltage", "hvac",
	"coolant", "bms", "state of charge", "soc", "degradation", "fire",
}

// loadAndChunkDocs loads all PDFs from docsDir within fsys, chunks them, and filters to battery-relevant chunks.
func loadAndChunkDocs(fsys fs.FS, docsDir string) ([]textChunk, error) {
	entries, err := fs.ReadDir(fsys, docsDir)
	if err != nil {
		return nil, fmt.Errorf("read docs dir %s: %w", docsDir, err)
	}

	var allChunks []textChunk
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".pdf") {
			continue
		}
		path := filepath.Join(docsDir, entry.Name())
		text, err := extractPDFTextFS(fsys, path)
		if err != nil {
			fmt.Printf("  Warning: could not read %s: %v\n", entry.Name(), err)
			continue
		}
		chunks := chunkText(text, 300, 30)
		allChunks = append(allChunks, chunks...)
	}

	// Filter to battery-relevant chunks only
	var filtered []textChunk
	for _, c := range allChunks {
		if isBatteryRelevant(c.text) {
			filtered = append(filtered, c)
		}
	}
	return filtered, nil
}

// extractPDFTextFS reads all text from a PDF file within an fs.FS.
func extractPDFTextFS(fsys fs.FS, path string) (string, error) {
	file, err := fsys.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	ra, ok := file.(interface {
		io.ReaderAt
	})
	if !ok {
		return "", fmt.Errorf("file does not support ReadAt: %s", path)
	}

	r, err := pdf.NewReader(ra, info.Size())
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteString(" ")
	}
	return sb.String(), nil
}

// chunkText splits text into overlapping chunks of approximately chunkSize words,
// with overlapSize words of overlap between consecutive chunks.
func chunkText(text string, chunkSize, overlapSize int) []textChunk {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var chunks []textChunk
	for i := 0; i < len(words); i += chunkSize - overlapSize {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, textChunk{text: chunk})
		if end == len(words) {
			break
		}
	}
	return chunks
}

func isBatteryRelevant(text string) bool {
	lower := strings.ToLower(text)
	for _, kw := range batteryKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
