FROM python:3.12-alpine AS builder

WORKDIR /app

RUN apk add --no-cache \
    build-base \
    linux-headers \
    musl-dev \
    zlib-dev \
    libffi-dev \
    postgresql-dev \
    openssl-dev \
    bash \
    curl \
    git

RUN pip install --upgrade pip
COPY requirements.txt .
RUN pip install --prefix=/install -r requirements.txt

# Set both PATH and PYTHONPATH
ENV PATH="/install/bin:${PATH}"
ENV PYTHONPATH="/install/lib/python3.12/site-packages:${PYTHONPATH}"

COPY . .
RUN pyinstaller --onefile src/main.py

FROM alpine:3.23
WORKDIR /app

RUN apk add --no-cache zlib libc6-compat libstdc++

COPY --from=builder /app/dist/main .

CMD ["./main"]
