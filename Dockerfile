FROM golang:1.18-alpine3.15 AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

WORKDIR /opt

RUN apk add --update curl make git libc-dev bash gcc linux-headers eudev-dev python3

RUN mkdir /opt/celinium

COPY . /opt/celinium/

ENV GOPROXY=https://goproxy.cn,direct

RUN cd /opt/celinium \
    && go install ./cmd/celiniumd

FROM alpine:3.15

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

RUN apk add --update bash

ENV PATH="/path/to/bash:$PATH"

COPY --from=builder /go/bin/celiniumd /usr/local/bin/

COPY ./docker/scripts/start_celi.sh /opt/start_celi.sh
COPY ./docker/scripts/helper.sh /opt/helper.sh

RUN chmod +x /opt/start_celi.sh
RUN chmod +x /opt/helper.sh

EXPOSE 26657 26656 1317 9090

ENTRYPOINT ["/opt/start_celi.sh"]
