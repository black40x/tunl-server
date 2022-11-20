# syntax=docker/dockerfile:1
FROM node:19-alpine AS ui

WORKDIR /ui/app

COPY ./ui/app /ui/app

RUN npm i

RUN npm run build

FROM golang:1.19.2-alpine

WORKDIR /app

COPY . .

COPY --from=ui /ui/app/build /app/ui/app/build

RUN go build -o /app/tunl-server ./cmd

COPY ./conf /app/conf

EXPOSE 8080 5000

CMD [ "/app/tunl-server" ]