FROM golang:1.12-alpine as builder

# Add Maintainer Info
LABEL maintainer="Sam Zhou <sam@mixmedia.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . /app

# Build the Go app
RUN go version \
 && export GO111MODULE=on \
 && export GOPROXY=https://goproxy.io \
 && go mod vendor \
 && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o html2pdf

######## Start a new stage from scratch #######
FROM debian:stretch-slim

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/html2pdf .
COPY --from=builder /app/web_root ./web_root
COPY --from=builder /app/config.json .
COPY --from=builder /app/render ./render
COPY --from=builder /app/font-conf ./font-conf

RUN set -x  \
# Install runtime dependencies
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        ca-certificates \
        bzip2 \
        libfontconfig \
        curl \
        git \
        fonts-droid-fallback \
        ttf-wqy-zenhei \
        ttf-wqy-microhei \
        fonts-arphic-ukai \
        fonts-arphic-uming \
        gettext-base \
# Install official PhantomJS release
 && mkdir /tmp/phantomjs \
 && curl -L https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-2.1.1-linux-x86_64.tar.bz2 \
        | tar -xj --strip-components=1 -C /tmp/phantomjs \
 && mv /tmp/phantomjs/bin/phantomjs /usr/local/bin/phantomjs \
 && ln -s /usr/local/bin/phantomjs /usr/bin/phantomjs \
# config font 
 && cp -r /app/font-conf/10-* /etc/fonts/conf.d/ \
 && fc-cache -fv \
# Install dumb-init (to handle PID 1 correctly).
# https://github.com/Yelp/dumb-init
 && curl -Lo /tmp/dumb-init.deb https://github.com/Yelp/dumb-init/releases/download/v1.1.3/dumb-init_1.1.3_amd64.deb \
 && dpkg -i /tmp/dumb-init.deb \
# Clean up
 && apt-get purge --auto-remove -y \
        curl git \
 && apt-get clean \
 && rm -rf /tmp/* /var/lib/apt/lists/* /app/*.gz  /app/font-conf

ENV WORKER=4 \
 HOST=0.0.0.0:4444 \
 ROOT=/app/web_root \
 TIMEOUT=60 \
 TLL=3600 \
 TZ=Asia/Hong_Kong

EXPOSE 4444

ENTRYPOINT ["dumb-init", "--"]

CMD  envsubst < /app/config.json > /app/temp.json \
 && /app/html2pdf -c /app/temp.json