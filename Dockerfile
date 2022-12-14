FROM golang:alpine AS builder

WORKDIR /go/src/app

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags "-extldflags \"-static\"" -o hollow main.go
RUN chmod +x hollow

FROM registry.cn-hangzhou.aliyuncs.com/bysir/alpine-shanghai:latest

COPY --from=builder /go/src/app/hollow /

ENTRYPOINT ["/hollow"]