FROM golang:1.18 AS builder

RUN pwd
RUN ls

RUN go build -v .

FROM registry.cn-hangzhou.aliyuncs.com/bysir/alpine-shanghai:latest

COPY --from=builder hollow /

ENTRYPOINT ["./hollow"]