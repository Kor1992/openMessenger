FROM golang:1.26.0-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/server ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=build /bin/server /bin/server

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1

CMD ["/bin/server"]
