FROM debian:jessie

ENV WORKER=4 HOST=0.0.0.0:4444 ROOT=/usr/local/html2pdf/web_root TIMEOUT=60

WORKDIR /root/src/github.com/mmhk/html2pdf/
COPY . .


RUN set -x  \
# Install runtime dependencies
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        ca-certificates \
        bzip2 \
        libfontconfig \
        curl \
        git \
        pdftk \
        fonts-droid \
        ttf-wqy-zenhei \
        ttf-wqy-microhei \
        fonts-arphic-ukai \
        fonts-arphic-uming \
        gettext-base \
# install go runtime
 && curl -O https://dl.google.com/go/go1.8.7.linux-amd64.tar.gz \
 && tar xvf go1.8.7.linux-amd64.tar.gz \
 && mv ./go /usr/local/go \
# build html2pdf
 && export GOPATH=/root \
 && export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin \
 && go get -v \
 && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o html2pdf . \
 && mkdir /usr/local/html2pdf \
 && mv web_root /usr/local/html2pdf/web_root \
 && mv render /usr/local/html2pdf/render \
 && mv config.json /usr/local/html2pdf/config.json \
 && mv html2pdf /usr/bin/html2pdf \
# Install official PhantomJS release
 && mkdir /tmp/phantomjs \
 && curl -L https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-2.1.1-linux-x86_64.tar.bz2 \
        | tar -xj --strip-components=1 -C /tmp/phantomjs \
 && mv /tmp/phantomjs/bin/phantomjs /usr/local/bin/phantomjs \
 && ln -s /usr/local/bin/phantomjs /usr/bin/phantomjs \
# Install dumb-init (to handle PID 1 correctly).
# https://github.com/Yelp/dumb-init
 && curl -Lo /tmp/dumb-init.deb https://github.com/Yelp/dumb-init/releases/download/v1.1.3/dumb-init_1.1.3_amd64.deb \
 && dpkg -i /tmp/dumb-init.deb \
# Clean up
 && apt-get purge --auto-remove -y \
        curl git \
 && apt-get clean \
 && rm -rf /tmp/* /var/lib/apt/lists/* \
 && rm -Rf /root/src \
 && rm -Rf /root/bin \
 && rm -Rf /root/pkg \
 && rm -Rf /usr/local/go 
 


EXPOSE 4444

ENTRYPOINT ["dumb-init"]

CMD  envsubst < /usr/local/html2pdf/config.json > /usr/local/html2pdf/config.json \
 && /usr/bin/html2pdf -c /usr/local/html2pdf/config.json