FROM golang:alpine as build
WORKDIR /build
RUN apk --no-cache add ca-certificates
COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -o /build/bin/app


FROM scratch
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt \
     /etc/ssl/certs/ca-certificates.crt
COPY --from=build /build/bin/app .
CMD ["./app"]
