FROM golang:1.18 AS builder

WORKDIR /go/src/app

COPY . .

RUN go build -v .

FROM registry.cn-hangzhou.aliyuncs.com/bysir/alpine-shanghai:latest

COPY --from=builder /go/src/app/hollow /

ENTRYPOINT ["./hollow"]