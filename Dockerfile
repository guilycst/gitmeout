ARG VERSION=dev

FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=${VERSION}" -o /gitmeout ./cmd/gitmeout

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata git

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /gitmeout /app/gitmeout

USER appuser

ENTRYPOINT ["/app/gitmeout"]