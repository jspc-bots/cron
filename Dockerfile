FROM golang:1.16.5-alpine as build

RUN apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates

ADD . /app
WORKDIR /app

RUN CGO_ENABLED=0 go build -a -ldflags="-s -w" -installsuffix cgo

FROM scratch

COPY --from=build /app/irc-timer-bot /irc-timer-bot
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo/Asia/Seoul /usr/share/zoneinfo/Asia/Seoul

CMD ["/irc-timer-bot"]
