FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY .  .
RUN go build -o contenthub .

FROM alpine:3.22

WORKDIR /server

COPY --from=builder /app/web ./web
COPY --from=builder /app/contenthub .

EXPOSE 9090

CMD [ "./contenthub" ]