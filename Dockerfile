FROM golang:1.15.5 as builder

WORKDIR /build

COPY . .

RUN go build -o goecho -mod=vendor -tags musl ./cmd/goecho

FROM golang:1.15.5

RUN mkdir /app 
WORKDIR /app

COPY --from=builder /build/goecho /app/goecho

COPY docker-entrypoint.sh .

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
