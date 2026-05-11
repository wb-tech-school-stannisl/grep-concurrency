package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wb-tech-school-stannisl/grep-concurrency/grep-node/internal/grep"
	pb "github.com/wb-tech-school-stannisl/grep-concurrency/grep-node/internal/proto/grep"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type grepNode struct {
	pb.UnimplementedGrepNodeServer
	nodeID string
}

func protoOptsToGrep(o *pb.Options) grep.Options {
	if o == nil {
		return grep.Options{}
	}
	return grep.Options{
		After:      int(o.After),
		Before:     int(o.Before),
		CountOnly:  o.CountOnly,
		IgnoreCase: o.IgnoreCase,
		Invert:     o.Invert,
		Fixed:      o.Fixed,
		LineNum:    o.LineNum,
	}
}

func (s *grepNode) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	if req.Pattern == "" {
		return nil, status.Error(codes.InvalidArgument, "pattern is required")
	}

	opts := protoOptsToGrep(req.Opts)

	result, err := grep.Run(req.Lines, req.Pattern, opts)
	if err != nil {
		return &pb.SearchResponse{
			ChunkIndex: req.ChunkIndex,
			NodeId:     s.nodeID,
			Ok:         false,
			Error:      err.Error(),
		}, nil
	}

	log.Printf("[%s] chunk=%d  matched=%d/%d lines",
		s.nodeID, req.ChunkIndex, result.Count, len(req.Lines))

	return &pb.SearchResponse{
		Matches:    result.Matches,
		Count:      int32(result.Count),
		ChunkIndex: req.ChunkIndex,
		NodeId:     s.nodeID,
		Ok:         true,
	}, nil
}

// Ping — healthcheck, оркестратор проверяет узел перед отправкой задачи
func (s *grepNode) Ping(_ context.Context, _ *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{NodeId: s.nodeID}, nil
}

func loggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	slog.Info("→", "RPC", info.FullMethod)

	resp, err := handler(ctx, req)
	if err != nil {
		slog.Info("✗", "RPC", info.FullMethod, "error", err)
	}

	return resp, err
}

func pingNode(addr string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost"+addr, //nolint:staticcheck
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: connect failed: %v\n", err)
		return 1
	}
	defer conn.Close()

	client := pb.NewGrepNodeClient(conn)
	resp, err := client.Ping(ctx, &pb.PingRequest{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: ping failed: %v\n", err)
		return 1
	}

	fmt.Printf("healthcheck: OK (node_id=%s)\n", resp.NodeId)
	return 0
}

func main() {
	port := flag.Int("port", 50051, "gRPC listen port")
	healthCheck := flag.Bool("health-check", false, "ping self and exit (for Docker HEALTHCHECK)")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	nodeID := fmt.Sprintf("node-localhost%s", addr)

	if *healthCheck {
		os.Exit(pingNode(addr))
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	pb.RegisterGrepNodeServer(server, &grepNode{nodeID: nodeID})

	slog.Info("grep-node listening", "grep-node", nodeID, "address", addr)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("serve error: %v", err)
		}
	}()

	<-ctx.Done()
	server.GracefulStop()
}
