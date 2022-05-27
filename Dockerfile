#build
FROM golang:1.18 AS build

RUN mkdir -p /home/app

COPY . /home/app

WORKDIR /home/app

RUN go build -v .

#deploy
FROM alpine:3.16.0

WORKDIR /

COPY --from=build /home/app/booking-app /booking-app

CMD ["./booking-app"]