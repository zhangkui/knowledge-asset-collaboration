FROM golang:1.26-bookworm AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/benzhi-server .

FROM alpine:3.20

RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/benzhi-server /usr/local/bin/benzhi-server

USER app
EXPOSE 8080
ENV HTTP_ADDR=:8080

ENTRYPOINT ["/usr/local/bin/benzhi-server"]
