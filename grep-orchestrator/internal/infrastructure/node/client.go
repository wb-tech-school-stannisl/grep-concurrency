package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/wb-tech-school-stannisl/grep-concurrency/grep-orchestrator/internal/config"
	"github.com/wb-tech-school-stannisl/grep-concurrency/grep-orchestrator/internal/input"
	pb "github.com/wb-tech-school-stannisl/grep-concurrency/grep-orchestrator/internal/proto/grep"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Call(cfg config.Config, addr string, lines []string, chunkIdx int32) (*pb.SearchResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", addr, err)
	}
	defer conn.Close()

	client := pb.NewGrepNodeClient(conn)

	resp, err := client.Search(ctx, &pb.SearchRequest{
		Pattern:    cfg.Pattern,
		Lines:      lines,
		ChunkIndex: chunkIdx,
		Opts: &pb.Options{
			After:      int32(cfg.After),
			Before:     int32(cfg.Before),
			CountOnly:  cfg.CountOnly,
			IgnoreCase: cfg.IgnoreCase,
			Invert:     cfg.Invert,
			Fixed:      cfg.Fixed,
			LineNum:    cfg.LineNum,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Search on %s: %w", addr, err)
	}
	return resp, nil
}

func SendToNodes(cfg config.Config, chunks [][]string) <-chan input.NodeResult {

	resultCh := make(chan input.NodeResult, len(cfg.Nodes))

	var wg sync.WaitGroup
	for i, node := range cfg.Nodes {
		wg.Add(1)

		go func(chunkIdx int, nodeAddr string) {
			defer wg.Done()

			chunk := []string{}
			if chunkIdx < len(chunks) {
				chunk = chunks[chunkIdx]
			}

			resp, err := Call(cfg, nodeAddr, chunk, int32(chunkIdx))
			resultCh <- input.NodeResult{
				Resp:   resp,
				Err:    err,
				NodeID: nodeAddr,
			}
		}(i, node)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return resultCh
}
