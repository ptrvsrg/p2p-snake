FROM golang:1.20-alpine as builder
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY ./cmd ./cmd
COPY ./internal ./internal
RUN go build -o /p2p-snake ./cmd/p2p-snake/main.go

FROM alpine:latest
COPY --from=builder p2p-snake /bin/p2p-snake
ENTRYPOINT /bin/p2p-snake -v -c ${CONFIG_FILE}