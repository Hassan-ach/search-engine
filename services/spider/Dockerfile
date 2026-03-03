FROM golang:tip-alpine3.23 AS go-builder

WORKDIR /app

RUN apk add --no-cache \
    build-base \
    just

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN just build

FROM alpine:3.23

WORKDIR /app

COPY --from=go-builder /app/spider .

CMD ["./spider"]
