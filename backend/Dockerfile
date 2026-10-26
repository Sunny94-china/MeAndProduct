# syntax=docker/dockerfile:1.6

FROM golang:1.25-alpine AS build
WORKDIR /src
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download || (sleep 3 && go mod download) || (sleep 5 && go mod download)
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -S -D -H app
WORKDIR /app
COPY --from=build /out/server /app/server
COPY migrations /app/migrations
USER app
ENV TZ=Asia/Shanghai
EXPOSE 8080
ENTRYPOINT ["/app/server"]
