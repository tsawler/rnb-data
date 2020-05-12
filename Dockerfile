#Download base image ubuntu 16.04
FROM alpine:3.11.6


EXPOSE 8080

WORKDIR /var/www

COPY places-linux places-linux


RUN chmod +x /var/www/places-linux


CMD /var/www/places-linux -addr ":8080" -dsn "homestead:secret@tcp($DOCKER_HOST:3306)/places?parseTime=true"
