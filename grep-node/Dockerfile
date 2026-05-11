FROM golang:1.26-alpine AS builder

RUN apk add --no-cache protobuf protobuf-dev curl

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /build

COPY proto/ ./proto/
RUN mkdir -p grep-node/internal/proto && \
    protoc \
      --go_out=grep-node/internal/proto \
      --go_opt=paths=source_relative \
      --go-grpc_out=grep-node/internal/proto \
      --go-grpc_opt=paths=source_relative \
      -I proto/grep \
      proto/grep/grep.proto

COPY grep-node/go.mod grep-node/go.sum* ./grep-node/
RUN cd grep-node && go mod download

COPY grep-node/ ./grep-node/

RUN cd grep-node && \
    CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w" \
      -o /bin/grep-node \
      ./cmd/grep-node

FROM scratch

COPY --from=builder /bin/grep-node /grep-node

EXPOSE 50051

ENTRYPOINT ["/grep-node"]
