# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o zrazabot main.go

FROM alpine:latest

WORKDIR /app

COPY zrazabot /app/zrazabot

RUN chmod +x /app/zrazabot

RUN mkdir -p /data

ENV DATA_DIR=/data

CMD ["./zrazabot"]