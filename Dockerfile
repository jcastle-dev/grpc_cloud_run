FROM golang:1.24.3-bookworm AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/server app/data.json ./
ENV DATA_JSON_URI=/app/data.json
EXPOSE 5000
CMD ["/app/server"]