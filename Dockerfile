FROM docker.1ms.run/library/golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev

# 配置 Go 模块代理
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o go-blog cmd/main.go

FROM docker.1ms.run/library/alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/go-blog .
COPY --from=builder /app/internal/template ./internal/template
COPY --from=builder /app/static ./static

RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./go-blog"]
