FROM alpine:3.11.6

EXPOSE 8080

WORKDIR /var/www

COPY places-linux places-linux

RUN chmod +x /var/www/places-linux

ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.7.3/wait /wait
RUN chmod +x /wait

CMD sh -c "/wait && /var/www/places-linux -addr \":8080\" -dsn \"root:secret@tcp($DB_HOST:3306)/places?parseTime=true\""
