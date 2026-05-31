# ---- 构建阶段 ----
FROM golang:1.24-alpine AS builder

WORKDIR /src

# 先拷贝 go.mod/go.sum 以利用层缓存
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 纯 Go 静态编译（Linux 下无需 cgo；默认 stub 行情源，gotdx 见 README 待办）
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/warden ./cmd/server

# ---- 运行阶段 ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app
COPY --from=builder /out/warden /app/warden
COPY --from=builder /src/config/config.yaml /app/config/config.yaml

EXPOSE 8080
ENTRYPOINT ["/app/warden"]
