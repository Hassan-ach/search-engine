FROM node:current-alpine3.23 AS builder

WORKDIR /app

COPY package*.json ./

RUN npm ci

COPY . .

RUN apk add --no-cache just

RUN just tailwind-build-prod
RUN just js-prod


FROM golang:tip-alpine3.23 AS go-builder

WORKDIR /app

RUN apk add --no-cache \
    build-base \
    aspell-dev \
    just

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY --from=builder /app/public ./public

RUN just build-emb

FROM alpine:3.23

WORKDIR /app

COPY --from=go-builder /app/boogle .

RUN apk add --no-cache aspell aspell-en

EXPOSE 1323

CMD ["./boogle"]
