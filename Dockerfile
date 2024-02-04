FROM golang:1.21.6-alpine
COPY ACE-Team-KKamJi /app
WORKDIR /app
RUN mkdir env && \
    echo DB_USER=admin >> env/.env && \
    echo DB_NAME=app >> env/.env && \
    go build -o main ./cmd/main.go && \
    chmod 755 main
ENTRYPOINT ["./main"]