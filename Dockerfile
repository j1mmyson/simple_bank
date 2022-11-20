# Build state
FROM golang:1.19.3-alpine3.16 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go


# Run stage
FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .

EXPOSE 3000
CMD [ "/app/main" ]
