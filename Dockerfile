# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/alp-drive ./cmd/alp-drive

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /bin/alp-drive ./alp-drive

EXPOSE 8080
ENTRYPOINT ["./alp-drive"]
