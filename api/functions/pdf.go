package functions

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

// extractTextFromPDF writes the PDF bytes to a temporary file and extracts all
// plain text. It first tries GetPlainText() which streams all pages at once;
// if that fails (e.g. "stream not present" on some PDF structures) it falls
// back to page-by-page extraction, skipping individual pages that error.
func ExtractTextFromPDF(pdfBytes []byte) (string, error) {
	tmp, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err = tmp.Write(pdfBytes); err != nil {
		tmp.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	// Close before opening with the pdf library so the file is fully flushed.
	if err = tmp.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	osFile, reader, err := pdf.Open(tmpName)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer osFile.Close()

	// Fast path: stream all pages at once via the reader-level API.
	if stream, streamErr := reader.GetPlainText(); streamErr == nil {
		var sb strings.Builder
		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			sb.WriteString(scanner.Text())
			sb.WriteString("\n")
		}
		if scanErr := scanner.Err(); scanErr != nil {
			return "", fmt.Errorf("error reading extracted text: %w", scanErr)
		}
		return sb.String(), nil
	}

	// Fallback: extract page by page. Some PDFs fail the stream-level call but
	// work fine when each page is processed individually.
	numPages := reader.NumPage()
	if numPages == 0 {
		return "", fmt.Errorf("PDF has no pages or could not be parsed")
	}

	var sb strings.Builder
	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		pageText, pageErr := page.GetPlainText(nil)
		if pageErr != nil {
			// Skip pages that individually fail; continue with the rest.
			continue
		}
		sb.WriteString(pageText)
	}

	return sb.String(), nil
}
