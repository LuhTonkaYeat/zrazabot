FROM alpine:latest

WORKDIR /app

# Копируем бинарник из локальной папки в контейнер
COPY zrazabot /app/zrazabot

# Даём права на выполнение
RUN chmod +x /app/zrazabot

# Запускаем бота
CMD ["./zrazabot"]