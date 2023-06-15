FROM golang:1.20.5-alpine AS builder

WORKDIR /go/src

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build main.go

# -----------------------------------------------------------------------------

FROM scratch

EXPOSE 8080

COPY --from=builder /go/src/main /

CMD ["/main"]
