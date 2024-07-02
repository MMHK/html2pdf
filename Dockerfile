FROM golang:1.19-alpine as builder

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

######## Start a new stage from bullseye #######
FROM chromedp/headless-shell:stable

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/html2pdf .
COPY --from=builder /app/web_root ./web_root
COPY --from=builder /app/font-conf ./font-conf

RUN set -x  \
# Install runtime dependencies
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        ca-certificates \
        dumb-init \
        gettext-base \
# config font
 && echo "deb http://deb.debian.org/debian/ bookworm main contrib" > /etc/apt/sources.list \
 && echo "deb-src http://deb.debian.org/debian/ bookworm main contrib" >> /etc/apt/sources.list \
 && echo "deb http://security.debian.org/ bookworm-security main contrib" >> /etc/apt/sources.list \
 && echo "deb-src http://security.debian.org/ bookworm-security main contrib" >> /etc/apt/sources.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        fontconfig \
        fonts-liberation \
        fonts-arphic-uming \
        fonts-arphic-ukai \
        fonts-droid-fallback \
        fonts-wqy-microhei \
        fonts-wqy-zenhei \
        fonts-noto \
        fonts-noto-cjk \
        fonts-unfonts-core \
        ttf-mscorefonts-installer \
 && cp -r /app/font-conf/10-* /etc/fonts/conf.d/ \
 && fc-cache -fv \
# Clean up
 && apt-get clean \
 && rm -rf /tmp/* /var/tmp/* /var/lib/apt/lists/* /app/*.gz  /app/font-conf

ENV WORKER=4 \
 LISTEN=0.0.0.0:4444 \
 WEB_ROOT=/app/web_root \
 TIMEOUT=60 \
 CHROME_PATH=/headless-shell/headless-shell \
 LOG_LEVEL=INFO \
 TZ=Asia/Hong_Kong

EXPOSE 4444

ENTRYPOINT ["dumb-init", "--"]

CMD  echo "{}" > /app/temp.json \
 && /app/html2pdf -c /app/temp.json