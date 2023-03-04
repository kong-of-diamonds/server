FROM golang:1.20-alpine3.17 as builder
WORKDIR /src
COPY . .
RUN go build -o build/app ./cmd/server/main.go

FROM alpine:3.17
WORKDIR /bin
COPY --from=builder /src/build/app ./
ENTRYPOINT [ "app" ]
