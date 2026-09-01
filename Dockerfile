ARG GO_VERSION=1.24

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:${GO_VERSION} AS builder

WORKDIR /src

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o /pd-slack ./cmd/main.go

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static

USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /pd-slack /pd-slack

LABEL org.opencontainers.image.source=https://github.com/imdevinc/pd-slack
LABEL org.opencontainers.image.description="Sync PagerDuty on-call users to a Slack user group."

ENTRYPOINT [ "/pd-slack" ]
