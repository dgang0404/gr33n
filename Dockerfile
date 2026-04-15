FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# SECURITY: Do NOT add `-tags dev` here — it enables auth bypass.
RUN CGO_ENABLED=0 go build -o /api ./cmd/api/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
