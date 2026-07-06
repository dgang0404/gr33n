package eval

import (
	"bufio"
	"os"
	"strings"
)

// ScrapeLogEvidence scans a log file for tool/eval markers (Phase 131 WS4).
func ScrapeLogEvidence(logPath, evalID, expectTool string) []string {
	logPath = strings.TrimSpace(logPath)
	if logPath == "" {
		return nil
	}
	f, err := os.Open(logPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var evidence []string
	seen := make(map[string]struct{})
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if evalID != "" && strings.Contains(line, evalID) {
			addEvidence(&evidence, seen, strings.TrimSpace(line))
		}
		if expectTool != "" && strings.Contains(line, "tool_id="+expectTool) {
			addEvidence(&evidence, seen, "tool_id="+expectTool)
		}
		if expectTool != "" && strings.Contains(line, expectTool) && strings.Contains(line, "guardian") {
			addEvidence(&evidence, seen, expectTool)
		}
	}
	return evidence
}

func addEvidence(out *[]string, seen map[string]struct{}, item string) {
	if item == "" {
		return
	}
	if _, ok := seen[item]; ok {
		return
	}
	seen[item] = struct{}{}
	*out = append(*out, item)
}
