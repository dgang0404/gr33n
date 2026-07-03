package fileattachutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const sniffLen = 512

// SniffAndValidate reads up to 512 bytes, detects MIME via magic bytes, and
// checks against allowed. Returns a reader replaying the sniff prefix + remainder.
func SniffAndValidate(r io.Reader, allowed map[string]struct{}) (mime string, body io.Reader, err error) {
	head := make([]byte, sniffLen)
	n, readErr := io.ReadFull(r, head)
	if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
		return "", nil, readErr
	}
	detected := strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(head[:n]), ";")[0]))
	if detected == "application/octet-stream" || detected == "" {
		return "", nil, fmt.Errorf("could not determine file type from content")
	}
	if _, ok := allowed[detected]; !ok {
		return "", nil, fmt.Errorf("unsupported file content type %q", detected)
	}
	return detected, io.MultiReader(bytes.NewReader(head[:n]), r), nil
}
