package input

import (
	"bufio"
	"log"
	"os"
	"sort"

	pb "github.com/wb-tech-school-stannisl/grep-concurrency/grep-orchestrator/internal/proto/grep"
)

type NodeResult struct {
	Resp   *pb.SearchResponse
	Err    error
	NodeID string
}

func SplitIntoChunks(lines []string, n int) [][]string {
	if n <= 0 {
		n = 1
	}
	size := (len(lines) + n - 1) / n // ceiling division
	chunks := make([][]string, 0, n)
	for i := 0; i < len(lines); i += size {
		end := i + size
		if end > len(lines) {
			end = len(lines)
		}
		chunks = append(chunks, lines[i:end])
	}
	return chunks
}

func CollectWithQuorum(
	resultCh <-chan NodeResult,
	quorum int,
	totalChunks int,
) ([]string, int, bool) {
	responses := make(map[int32]*pb.SearchResponse, totalChunks)

	successCount := 0
	failCount := 0

	for nr := range resultCh {
		if nr.Err != nil || (nr.Resp != nil && !nr.Resp.Ok) {
			errMsg := ""
			if nr.Err != nil {
				errMsg = nr.Err.Error()
			} else if nr.Resp != nil {
				errMsg = nr.Resp.Error
			}
			log.Printf("✗ node %s failed: %s", nr.NodeID, errMsg)
			failCount++
			continue
		}

		log.Printf("✓ node %s chunk=%d matches=%d",
			nr.NodeID, nr.Resp.ChunkIndex, len(nr.Resp.Matches))

		responses[nr.Resp.ChunkIndex] = nr.Resp
		successCount++
	}

	log.Printf("quorum: %d/%d successful (need %d)", successCount, totalChunks, quorum)

	if successCount < quorum {
		return nil, 0, false
	}

	return MergeOrdered(responses)
}

// mergeOrdered собирает строки из всех чанков в правильном порядке
func MergeOrdered(responses map[int32]*pb.SearchResponse) ([]string, int, bool) {
	// сортируем ключи
	keys := make([]int, 0, len(responses))
	for k := range responses {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	var allMatches []string
	totalCount := 0

	for _, k := range keys {
		resp := responses[int32(k)]
		allMatches = append(allMatches, resp.Matches...)
		totalCount += int(resp.Count)
	}

	return allMatches, totalCount, true
}

func ReadInput(args []string) []string {
	var reader *bufio.Scanner

	if len(args) > 0 {
		// читаем из файла
		f, err := os.Open(args[0])
		if err != nil {
			log.Fatalf("open file %q: %v", args[0], err)
		}
		defer f.Close()
		reader = bufio.NewScanner(f)
	} else {
		// читаем из stdin
		reader = bufio.NewScanner(os.Stdin)
	}

	buf := make([]byte, 0, 64*1024)
	reader.Buffer(buf, 10*1024*1024)

	var lines []string
	for reader.Scan() {
		lines = append(lines, reader.Text())
	}
	if err := reader.Err(); err != nil {
		log.Fatalf("read input: %v", err)
	}
	return lines
}
