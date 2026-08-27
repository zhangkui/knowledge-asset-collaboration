# Keep the Go compiler and Node.js toolchain in one image for build and runtime.
FROM golang:1.26-bookworm

# The frontend build requires Node.js. Node 20 is the project delivery baseline.
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Download Go modules before copying the remaining source for better build caching.
COPY go.mod ./
RUN go mod download

# Install frontend dependencies before copying the complete frontend source.
COPY web/package.json web/package-lock.json ./web/
RUN cd web && npm ci --no-audit --no-fund

COPY . .

# Build the backend executable and the Vite frontend in the same image.
RUN go build -trimpath -ldflags='-s -w' -o /usr/local/bin/benzhi-server . \
    && cd web && npm run build

EXPOSE 8080
ENV HTTP_ADDR=:8080
ENV WEB_DIR=/app/web/dist

CMD ["/usr/local/bin/benzhi-server"]
