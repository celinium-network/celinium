ARG IMG_TAG=latest

# Compile the celiniumd binary
FROM golang:1.18-alpine AS celiniumd-builder
WORKDIR /src/app/
COPY go.mod go.sum* ./

ENV GOPROXY=https://goproxy.cn,direct
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

RUN go mod download
COPY . .
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3
RUN apk add --no-cache $PACKAGES
RUN CGO_ENABLED=0 go install ./cmd/celiniumd

# Add to a distroless container
FROM distroless.dev/static:$IMG_TAG
ARG IMG_TAG
COPY --from=celiniumd-builder /go/bin/celiniumd /usr/local/bin/
EXPOSE 26656 26657 1317 9090

ENTRYPOINT ["celiniumd", "start"]
