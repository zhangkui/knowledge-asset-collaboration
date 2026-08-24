FROM golang:1.26 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/server .
FROM debian:bookworm-slim
COPY --from=build /out/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
