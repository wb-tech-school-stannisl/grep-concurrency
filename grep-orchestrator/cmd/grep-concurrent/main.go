package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/wb-tech-school-stannisl/grep-concurrency/grep-orchestrator/internal/config"
	"github.com/wb-tech-school-stannisl/grep-concurrency/grep-orchestrator/internal/infrastructure/node"
	"github.com/wb-tech-school-stannisl/grep-concurrency/grep-orchestrator/internal/input"
)

func main() {
	config := config.LoadConfig()

	lines := input.ReadInput(flag.Args())

	if len(lines) == 0 {
		fmt.Fprintln(os.Stderr, "no input lines")
		os.Exit(1)
	}

	chunks := input.SplitIntoChunks(lines, len(config.Nodes))

	log.Printf("input: %d lines → %d chunks across %d nodes (quorum=%d)",
		len(lines), len(chunks), len(config.Nodes), config.Quorum)

	// запускаем конкурентные запросы к узлам
	results := node.SendToNodes(config, chunks)

	// проверяем кворум и собираем ответ
	matches, count, ok := input.CollectWithQuorum(results, config.Quorum, len(chunks))
	if !ok {
		fmt.Fprintln(os.Stderr, "quorum not reached: too many nodes failed")
		os.Exit(2)
	}

	if config.CountOnly {
		fmt.Println(count)
		return
	}
	for _, line := range matches {
		fmt.Println(line)
	}
}
