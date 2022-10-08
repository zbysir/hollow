FROM registry.cn-hangzhou.aliyuncs.com/bysir/alpine-shanghai:latest

COPY hollow /

ENTRYPOINT ["./hollow"]