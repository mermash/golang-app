#build
FROM golang:1.18.2-alpine3.16 AS build

RUN mkdir -p /home/app

COPY . /home/app

WORKDIR /home/app

#RUN CGO_ENABLED=0 go build -v -o /home/app/booking-app .
RUN go build -v -o /home/app/booking-app .

#deploy
FROM alpine:3.16.0

WORKDIR /go/bin

COPY --from=build /home/app/booking-app ./

# ENV CONFNAME="GO Conference"
# ENV DB_HOST=mysql-db
# ENV DB_PORT=3306
# ENV DB_USER=root
# ENV DB_PASSWORD=root
# ENV DB_DB=my_golang_app

CMD ["./booking-app"]