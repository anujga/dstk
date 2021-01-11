# Build the manager binary
FROM golang:1.14-buster as builder
RUN curl -sL https://taskfile.dev/install.sh | sh

COPY scripts/depends.sh /
COPY go.mod /

RUN bash /depends.sh

