FROM golang:1.11.1-alpine AS server-build
ENV GO111MODULE=on
RUN apk add --no-cache git
WORKDIR /src
COPY ./go.* ./
RUN go mod download
COPY ./*.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /cmkp

FROM node:10.10.0-alpine as client-build
WORKDIR /frontend
COPY ./frontend/package.json ./frontend/yarn.lock ./
RUN yarn install
COPY ./frontend .
RUN yarn build

# runtime image
FROM alpine:3.8
RUN apk add --no-cache ca-certificates openssl
ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz
EXPOSE 3000
COPY --from=server-build /cmkp /cmkp
COPY --from=client-build /frontend/dist /static
ENTRYPOINT ["/cmkp"]
