FROM golang:1.21.0 as builder
RUN mkdir /app

WORKDIR /app

COPY ./ ./

RUN go mod tidy

RUN go build

FROM ubuntu:22.04 as runner

COPY --from=builder /app/go-test ./

ENTRYPOINT ["./go-test"]