FROM golang:1.24

WORKDIR /app

COPY api/go.mod api/go.sum ./api/

WORKDIR /app/api
RUN go mod download -x

WORKDIR /app
COPY api ./api
COPY migrations ./migrations

WORKDIR /app/api
RUN go build -o main .

EXPOSE 8888

CMD ["./main"]
