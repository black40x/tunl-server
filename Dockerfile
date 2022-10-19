# syntax=docker/dockerfile:1

FROM golang:1.19.2-alpine

WORKDIR /tunl

COPY . .

RUN go build -o /tunl/tunl-server ./cmd

COPY ./conf /tunl/conf

EXPOSE 8080 5000

CMD [ "/tunl/tunl-server" ]