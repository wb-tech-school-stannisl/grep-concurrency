build:
	cd grep-orchestrator/cmd/grep-concurrent && go build -o ../../../bin/main_grep .

start-nodes:
	docker compose up -d

test-3-out-of-3-available: build start-nodes
	 cat test-document | ./bin/main_grep -e "test" -nodes localhost:50051,localhost:50052,localhost:50053 -quorum 2

test-2-out-of-3-available: build start-nodes
	 cat test-document | ./bin/main_grep -e "test" -nodes localhost:50051,localhost:50052,localhost:50054 -quorum 2

test-1-out-of-3-available: build start-nodes
	 cat test-document | ./bin/main_grep -e "test" -nodes localhost:50051,localhost:50055,localhost:50054 -quorum 2
