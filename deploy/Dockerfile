# Build the manager binary
FROM golang:1.14 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum

COPY examples/ examples/
COPY pkg pkg/
COPY vendor vendor/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o counter github.com/anujga/dstk/examples/persistent-counters

FROM golang:1.14.2-alpine
WORKDIR /
COPY --from=builder /workspace/counter .
COPY --from=builder /workspace/examples/persistent-counters/config.yaml .
ENTRYPOINT ["/counter", "-conf", "/"]
