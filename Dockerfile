# Build the binary
FROM --platform=$BUILDPLATFORM golang:1.23 AS builder

ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY cmd cmd/
COPY docs docs/
COPY internal internal/
COPY pkg pkg/
COPY config.yaml config.yaml

# Build
RUN go generate ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH="$TARGETARCH" go build -a -o time_keeper main.go

# Use distroless as minimal base image to package hanko binary
# See https://github.com/GoogleContainerTools/distroless for details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/time_keeper .
USER 65532:65532

ENTRYPOINT ["/time_keeper"]
