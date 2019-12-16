FROM alpine
RUN apk add --update ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/*
ADD mikudos-message-pusher-srv /mikudos-message-pusher-srv
WORKDIR /
ENTRYPOINT [ "/mikudos-message-pusher-srv" ]
