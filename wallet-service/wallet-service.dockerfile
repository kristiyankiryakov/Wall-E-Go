FROM alpine:latest

RUN mkdir /app

COPY walletApp /app

CMD [ "/app/walletApp" ]