# syntax=docker/dockerfile:1.7
FROM golang:1.25-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
    go build -trimpath \
      -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
      -o /out/dogelytics ./cmd/dogelytics

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /out/dogelytics /dogelytics

USER 65532:65532
EXPOSE 4420 4421 4422
HEALTHCHECK --interval=15s --timeout=5s --start-period=15s --retries=5 \
  CMD ["/dogelytics", "healthcheck", "--url", "http://127.0.0.1:4420/readyz"]
ENTRYPOINT ["/dogelytics"]
CMD ["serve"]
