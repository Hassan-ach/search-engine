FROM rust:alpine3.23 AS builder

RUN apk add --no-cache musl-dev musl-utils build-base

WORKDIR /app

COPY . .

ENV SQLX_OFFLINE=true
ENV RUSTFLAGS="-C target-cpu=native"

RUN rustup target add x86_64-unknown-linux-musl
RUN cargo build --release --target x86_64-unknown-linux-musl

FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/target/x86_64-unknown-linux-musl/release/indexer .
CMD ["./indexer"]
