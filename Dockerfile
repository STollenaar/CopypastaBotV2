FROM golang:1.18.0-alpine

RUN apk update && apk add git

# Create app directory
WORKDIR /usr/src/app

COPY . .

RUN go get .

CMD go run .
