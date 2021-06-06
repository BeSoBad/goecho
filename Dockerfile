FROM golang:1.15.5 as builder

WORKDIR /build

COPY . .

RUN go build -o echo ./cmd/echo

FROM golang:1.15.5

WORKDIR /app

COPY --from=builder /build/echo /app/echo

COPY docker-entrypoint.sh .

EXPOSE 7

ENTRYPOINT ["/app/docker-entrypoint.sh"]
