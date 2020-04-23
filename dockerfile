FROM golang:1.13 as builder

WORKDIR /go/src/app
COPY . .

RUN GOOS=linux CGO_ENABLED=0 go build -o jwt cmd/main.go

FROM alpine:3.8

WORKDIR /root
COPY --from=builder /go/src/app/jwt .

EXPOSE 8080

RUN chmod 0755 jwt
CMD [ "./jwt" ]
