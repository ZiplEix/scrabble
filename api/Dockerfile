FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download -x

COPY . .

RUN go build -o main .

EXPOSE 8888

CMD ["./main"]
