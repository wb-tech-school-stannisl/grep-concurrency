package grep

import (
	"fmt"
	"regexp"
	"strings"
)

type Options struct {
	After      int
	Before     int
	CountOnly  bool
	IgnoreCase bool
	Invert     bool
	Fixed      bool
	LineNum    bool
}

type Result struct {
	Matches []string
	Count   int
}

func Run(lines []string, pattern string, opts Options) (Result, error) {
	var re *regexp.Regexp
	if !opts.Fixed {
		flags := ""
		if opts.IgnoreCase {
			flags = "(?i)"
		}
		var err error
		re, err = regexp.Compile(flags + pattern) // ✅ = вместо :=
		if err != nil {
			return Result{}, fmt.Errorf("invalid pattern: %w", err)
		}
	}

	// поиск совпадений
	matches := make([]bool, len(lines))
	for i, line := range lines {
		var match bool
		if opts.Fixed {
			if opts.IgnoreCase {
				match = strings.Contains(strings.ToLower(line), strings.ToLower(pattern))
			} else {
				match = strings.Contains(line, pattern)
			}
		} else {
			match = re.MatchString(line)
		}
		if opts.Invert {
			match = !match
		}
		matches[i] = match
	}

	// контекст (Before/After)
	toPrint := make([]bool, len(lines))
	for i, match := range matches {
		if match {
			start := max(0, i-opts.Before)
			end := min(len(lines)-1, i+opts.After)
			for j := start; j <= end; j++ {
				toPrint[j] = true
			}
		}
	}

	// ✅ Собираем результат в структуру, не печатаем напрямую
	result := Result{}
	for i, p := range toPrint {
		if p {
			if opts.LineNum {
				result.Matches = append(result.Matches, fmt.Sprintf("%d:%s", i+1, lines[i]))
			} else {
				result.Matches = append(result.Matches, lines[i])
			}
		}
	}

	// считаем только прямые совпадения (без контекста)
	for _, m := range matches {
		if m {
			result.Count++
		}
	}

	return result, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
