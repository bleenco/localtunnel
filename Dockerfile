FROM alpine:latest

LABEL maintainer="Jan Kuri <jan@bleenco.com>"

ENV DOMAIN=bleenco.space
ENV SECURE=true

RUN apk --no-cache add tini curl ca-certificates
COPY ./build/linux_amd64/lt-server /lt-server
EXPOSE 1234

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s CMD curl -f http://localhost:1234 || exit 1

ENTRYPOINT ["/sbin/tini", "--"] CMD curl -f http://localhost:1234 || exit 1
CMD /lt-server -p 1234 -d $DOMAIN -s=$SECURE
