FROM alpine:latest

RUN mkdir /app

COPY compApp /app

CMD [ "/app/compApp"]