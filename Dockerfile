FROM golang:1.17-alpine AS builder
ENV GO111MODULE=on
RUN apk add -U --no-cache ca-certificates
WORKDIR /go-bandit-api
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/go-bandit-api

FROM scratch
WORKDIR /bin/go-bandit-api/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go-bandit-api/config.conf.docker ./config.conf
COPY --from=builder /bin/go-bandit-api ./go-bandit-api
CMD ["./go-bandit-api"]
