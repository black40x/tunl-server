# syntax=docker/dockerfile:1

FROM golang:1.19.2-alpine

WORKDIR /app

COPY . .

RUN go build -o /app/tunl-server ./cmd

COPY ./conf /app/conf

EXPOSE 8080 5000

CMD [ "/app/tunl-server" ]