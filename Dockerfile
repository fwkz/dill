FROM golang:1.19-alpine AS builder
WORKDIR /tmp/dill
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o dist/dill cmd/dill/main.go

FROM alpine:3.17.2
RUN apk update && apk --no-cache upgrade
COPY --from=builder /tmp/dill/dist/dill /usr/local/bin

ENTRYPOINT ["/usr/local/bin/dill"]