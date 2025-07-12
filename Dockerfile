FROM golang:1.21-alpine

WORKDIR /app

COPY ./src/go.mod ./
RUN go mod download

COPY ./src .

RUN go build -o server .

EXPOSE 5000

CMD ["./server"]