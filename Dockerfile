FROM golang:1.15.5 as builder

WORKDIR /build

COPY . .

RUN go build -o echo -mod=vendor -tags musl ./cmd/echo

FROM golang:1.15.5

RUN mkdir /app 
WORKDIR /app

COPY --from=builder /build/echo /app/echo

COPY docker-entrypoint.sh .

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
