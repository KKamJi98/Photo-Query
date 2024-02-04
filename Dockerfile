FROM golang:1.21.6-alpine
COPY ./* /picture-backend
WORKDIR /picture-backend
RUN set -x && \
    mkdir env && \
    echo DB_USER=admin >> ./env/.env && \
    echo DB_NAME=app >> ./env/.env && \
    go build -o main ./cmd/main.go && \
    chmod 755 main
ENTRYPOINT ["./main"]