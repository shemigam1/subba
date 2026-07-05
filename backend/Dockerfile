# Multi-stage build for the three engine binaries. All targets share one
# builder layer, so `--target api|worker|scheduler` reuses the same compile
# cache. Static binaries (CGO off) on distroless: ~25MB images, no shell,
# minimal attack surface.

FROM golang:1.25-alpine AS builder
WORKDIR /src

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/scheduler ./cmd/scheduler

FROM gcr.io/distroless/static-debian12:nonroot AS api
COPY --from=builder /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]

FROM gcr.io/distroless/static-debian12:nonroot AS worker
COPY --from=builder /out/worker /worker
ENTRYPOINT ["/worker"]

FROM gcr.io/distroless/static-debian12:nonroot AS scheduler
COPY --from=builder /out/scheduler /scheduler
ENTRYPOINT ["/scheduler"]

# `all` bundles the three binaries in one image so a single ECR repo/tag can run
# api, worker, and scheduler as separate compose services (override entrypoint per
# service). Used by the Elastic Beanstalk CI/CD pipeline.
FROM gcr.io/distroless/static-debian12:nonroot AS all
COPY --from=builder /out/api /api
COPY --from=builder /out/worker /worker
COPY --from=builder /out/scheduler /scheduler
EXPOSE 8080
ENTRYPOINT ["/api"]
