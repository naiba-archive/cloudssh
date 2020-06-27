FROM golang:alpine AS binarybuilder
# Install build deps
RUN apk --no-cache --no-progress add --virtual build-deps build-base git linux-pam-dev
WORKDIR /go/src/github.com/naiba/cloudssh/
COPY . .
RUN cd cmd/server \
    && go build -ldflags="-s -w"

FROM alpine:latest
RUN echo http://dl-2.alpinelinux.org/alpine/edge/community/ >> /etc/apk/repositories \
  && apk --no-cache --no-progress add \
    tzdata
# Copy binary to container
WORKDIR /cloudssh
ADD resource resource
COPY --from=binarybuilder /go/src/github.com/naiba/cloudssh/cmd/server/server ./cloudssh

# Configure Docker Container
VOLUME ["/cloudssh/data"]
EXPOSE 8000
CMD ["/cloudssh/cloudssh","-conf","/cloudssh/data/config.json"]
