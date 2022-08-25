FROM golang:1.18.0-alpine

ARG ARCH


# Create app directory
WORKDIR /usr/src/app

COPY CopypastaBotV2 .

CMD ./CopypastaBotV2