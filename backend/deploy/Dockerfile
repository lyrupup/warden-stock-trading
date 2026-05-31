# ---- 构建阶段 ----
# go.mod 要求 Go 1.26（gotdx 依赖）；GOTOOLCHAIN=auto 兜底自动拉取匹配工具链。
FROM golang:1.26-alpine AS builder

WORKDIR /src

ENV GOTOOLCHAIN=auto

# 先拷贝 go.mod/go.sum 以利用层缓存
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 纯 Go 静态编译（Linux 下无需 cgo）。
# 启用 -tags gotdx 接入真实通达信行情源，与 compose 的 MARKET_PROVIDER=gotdx 对应。
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -tags gotdx -trimpath -ldflags="-s -w" -o /out/warden ./cmd/server

# ---- 运行阶段 ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app
COPY --from=builder /out/warden /app/warden
COPY --from=builder /src/config/config.yaml /app/config/config.yaml

EXPOSE 8080
ENTRYPOINT ["/app/warden"]
