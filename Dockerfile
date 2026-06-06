# Build the binary
FROM --platform=$BUILDPLATFORM golang:1.25 AS builder

ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy source (templ-generated *_templ.go files must be committed before building)
COPY main.go main.go
COPY cmd cmd/
COPY docs docs/
COPY internal internal/
COPY pkg pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH="$TARGETARCH" go build -a -o time_keeper main.go

# Minimal runtime image
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/time_keeper .
# Static assets (Tailwind CSS output — run `task tailwind` before building)
COPY static static/
USER 65532:65532

ENTRYPOINT ["/time_keeper"]
