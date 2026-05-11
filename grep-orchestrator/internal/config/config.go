package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Pattern   string
	Nodes     []string
	Quorum    int
	Timeout   time.Duration
	ChunkSize int

	After      int
	Before     int
	CountOnly  bool
	IgnoreCase bool
	Invert     bool
	Fixed      bool
	LineNum    bool
}

func LoadConfig() Config {
	pattern := flag.String("e", "", "search pattern (required)")
	nodesStr := flag.String("nodes", "localhost:50051,localhost:50052,localhost:50053",
		"comma-separated list of node addresses")
	quorum := flag.Int("quorum", 0,
		"min successful nodes (default: majority)")
	timeout := flag.Duration("timeout", 5*time.Second, "per-node timeout")

	after := flag.Int("A", 0, "lines of context after match")
	before := flag.Int("B", 0, "lines of context before match")
	countOnly := flag.Bool("c", false, "print count only")
	ignoreCase := flag.Bool("i", false, "ignore case")
	invert := flag.Bool("v", false, "invert match")
	fixed := flag.Bool("F", false, "fixed string (no regexp)")
	lineNum := flag.Bool("n", false, "print line numbers")

	flag.Parse()

	if *pattern == "" {
		fmt.Fprintln(os.Stderr, "usage: mygrep -e PATTERN [-nodes ...] [-quorum N] [file]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	nodes := strings.Split(*nodesStr, ",")
	for i := range nodes {
		nodes[i] = strings.TrimSpace(nodes[i])
	}

	q := *quorum
	if q == 0 {
		q = len(nodes)/2 + 1
	}

	return Config{
		Pattern:    *pattern,
		Nodes:      nodes,
		Quorum:     q,
		Timeout:    *timeout,
		After:      *after,
		Before:     *before,
		CountOnly:  *countOnly,
		IgnoreCase: *ignoreCase,
		Invert:     *invert,
		Fixed:      *fixed,
		LineNum:    *lineNum,
	}
}
