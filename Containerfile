FROM golang:1.23-alpine AS go-upx

RUN apk add --no-cache git upx tzdata

ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"


FROM go-upx AS builder

WORKDIR /app

ENV CGO_ENABLED=0

COPY main.go go.mod go.sum  /app/

RUN go mod download && \
    go mod verify && \
    go get -d ./...

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" .

RUN upx qpsync


FROM scratch

COPY --from=builder /etc/passwd             /etc/passwd
COPY --from=builder /etc/group              /etc/group
COPY --from=builder /usr/share/zoneinfo     /usr/share/zoneinfo
COPY --from=builder /app/qpsync             /app/

USER appuser
WORKDIR /app

ENTRYPOINT ["/app/qpsync"]
