FROM golang:1.24.4 AS builder

COPY . /src

WORKDIR /src

RUN GOPROXY=https://goproxy.cn go build -o ./bin/godo ./cmd/

FROM debian:stable-slim

ENV TZ=Asia/Shanghai

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app

WORKDIR /app

EXPOSE 8080

VOLUME /config
VOLUME /app/uploads
VOLUME /app/logs

CMD ["./godo", "-conf", "/config/config.yaml"]
