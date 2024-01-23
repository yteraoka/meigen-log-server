FROM golang:1.21.6 AS builder

ARG version

WORKDIR /build
COPY main.go go.mod go.sum ./
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o meigen -ldflags "-w -s -X main.version=${version:-dev}"

# hadolint ignore=DL3006
FROM gcr.io/distroless/base
WORKDIR /
COPY --from=builder /build/meigen /meigen
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/meigen"]
