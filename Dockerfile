FROM golang:1.12.5-alpine AS server-build
ENV GO111MODULE=on
RUN apk add --no-cache git
WORKDIR /go/src/github.com/wtks/cmkp/backend
COPY ./backend/go.* ./
RUN go mod download
COPY ./backend .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /cmkp

# runtime image
FROM alpine:3.8
RUN apk add --no-cache ca-certificates openssl
ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz
EXPOSE 3000
COPY --from=server-build /cmkp /cmkp
ENTRYPOINT ["/cmkp"]
