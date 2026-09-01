package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// ParseSelection maps user input to 0-based indexes into ssids.
// Accepts "all", comma-separated numbers, or whitespace-separated numbers.
func ParseSelection(input string, count int) ([]int, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, fmt.Errorf("select at least one network")
	}
	if strings.EqualFold(trimmed, "all") {
		indexes := make([]int, count)
		for i := range indexes {
			indexes[i] = i
		}
		return indexes, nil
	}

	fields := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	if len(fields) == 0 {
		return nil, fmt.Errorf("select at least one network")
	}

	seen := make(map[int]struct{}, len(fields))
	indexes := make([]int, 0, len(fields))
	for _, field := range fields {
		n, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("invalid selection %q", field)
		}
		if n < 1 || n > count {
			return nil, fmt.Errorf("selection %d is out of range (1-%d)", n, count)
		}
		idx := n - 1
		if _, ok := seen[idx]; ok {
			continue
		}
		seen[idx] = struct{}{}
		indexes = append(indexes, idx)
	}
	return indexes, nil
}

// ReadLine reads a single line from r, without the trailing newline.
func ReadLine(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 64*1024)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("read input: %w", err)
		}
		return "", fmt.Errorf("read input: no line")
	}
	return scanner.Text(), nil
}
