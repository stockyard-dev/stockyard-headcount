FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go mod download && CGO_ENABLED=0 go build -o headcount ./cmd/headcount/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/headcount .
ENV PORT=9160 DATA_DIR=/data
EXPOSE 9160
CMD ["./headcount"]
