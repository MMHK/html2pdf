FROM golang:1.20.14-bullseye as builder

# Add Maintainer Info
LABEL maintainer="Sam Zhou <sam@mixmedia.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . /app

# Build the Go app
RUN apt-get update \
 && apt-get install -y --no-install-recommends curl git ca-certificates \
 && go version \
 && export GOPROXY=https://proxy.golang.org,direct \
 && go mod tidy \
 && go mod vendor \
 && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o html2pdf \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

######## Start a new stage from bullseye #######
FROM chromedp/headless-shell:stable

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/html2pdf .
COPY --from=builder /app/web_root ./web_root
COPY --from=builder /app/font-conf ./font-conf

# 正確啟用 contrib（Trixie 使用 DEB822 格式，修改 debian.sources）
RUN sed -i 's/Components: main$/Components: main contrib/' /etc/apt/sources.list.d/debian.sources \
    && apt-get update

    # 安裝必要工具（下載字體需要）
RUN apt-get install -y --no-install-recommends \
        wget \
        ca-certificates \
        cabextract \
        xfonts-utils

# 安裝其他字體包（這些沒問題）
RUN apt-get install -y --no-install-recommends \
        fontconfig \
        fonts-liberation \
        fonts-arphic-uming \
        fonts-arphic-ukai \
        fonts-droid-fallback \
        fonts-wqy-microhei \
        fonts-wqy-zenhei \
        fonts-noto \
        fonts-noto-cjk \
        fonts-unfonts-core

RUN set -x  \
# Install runtime dependencies
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        ca-certificates \
        dumb-init \
        gettext-base \
 && cp -r /app/font-conf/10-* /etc/fonts/conf.d/ \
 && fc-cache -fv \
# Clean up
 && apt-get clean \
 && rm -rf /tmp/* /var/tmp/* /var/lib/apt/lists/* /app/*.gz  /app/font-conf

ENV WORKER=4 \
 LISTEN=0.0.0.0:4444 \
 WEB_ROOT=/app/web_root \
 TIMEOUT=60 \
 PDF_AUTHOR=driver.com.hk \
 PDF_CREATOR=HTML2PDF \
 PDF_KEYWORDS=driver.com.hk,html2pdf \
 PDF_SUBJECT=PDFDocument \
 CHROME_PATH=/headless-shell/headless-shell \
 LOG_LEVEL=INFO \
 CLEANER_PERIOD=1800 \
 CLEANER_FILE_AGE_LIMIT=86400 \
 TZ=Asia/Hong_Kong

EXPOSE 4444

ENTRYPOINT ["dumb-init", "--"]

CMD  echo "{}" > /app/temp.json \
 && /app/html2pdf -c /app/temp.json