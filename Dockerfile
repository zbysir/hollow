FROM golang:1.18 as builder

RUN go build -v .

FROM registry.cn-hangzhou.aliyuncs.com/bysir/alpine-shanghai:latest

COPY --from=builder hollow /

ENTRYPOINT ["./hollow"]