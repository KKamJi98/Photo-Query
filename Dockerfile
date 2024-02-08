FROM golang:1.22.0-alpine
WORKDIR /picture-backend
COPY ./main .
# RUN set -x && \
#     mkdir env && \
#     echo DB_USER=admin >> /picture-backend/env/.env && \
#     echo DB_NAME=app >> /picture-backend/env/.env && \
#     go mod tidy && \
#     go build -o main /picture-backend/cmd/main.go && \
#     chmod 755 ./main
ENTRYPOINT ["./main"]