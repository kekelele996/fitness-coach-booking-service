# 评测用镜像：构建并启动真实 HTTP 服务。
FROM golang:1.22
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /usr/local/bin/server ./cmd/server
EXPOSE 8080
CMD ["/usr/local/bin/server"]
