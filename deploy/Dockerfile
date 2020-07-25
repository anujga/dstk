# Build the manager binary
FROM golang:1.14-buster as builder

WORKDIR /

RUN curl -sL https://taskfile.dev/install.sh | sh

WORKDIR /workspace
COPY . .

RUN /bin/task apps

FROM ubuntu:20.04
WORKDIR /
COPY --from=builder /workspace/bin /dstk
#COPY --from=builder /workspace/examples/persistent-counters/config.yaml .
#ENTRYPOINT ["/bin/bash"]
#ENTRYPOINT ["/counter", "-conf", "/"]
