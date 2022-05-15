FROM golang:1.17 as builder

WORKDIR /app

COPY . .

RUN go build -o shareProcessNS main.go

FROM alpine

WORKDIR /app

COPY --from=builder /app/shareProcessNS .

CMD ["./shareProcessNS"]