FROM alpine:latest

RUN mkdir /app && apk add --no-cache curl

COPY brokerApp /app

CMD [ "/app/brokerApp" ]