FROM alpine:latest
RUN echo http://dl-2.alpinelinux.org/alpine/edge/community/ >> /etc/apk/repositories \
  && apk --no-cache --no-progress add \
  tzdata
# Copy binary to container
WORKDIR /cloudssh
ADD resource resource
ADD server cloudssh

# Configure Docker Container
VOLUME ["/cloudssh/data"]
EXPOSE 8000
CMD ["/cloudssh/cloudssh","-conf","/cloudssh/data/config.json"]
