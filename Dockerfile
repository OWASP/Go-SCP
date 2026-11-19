FROM node:12.11.1-slim

# Debian stretch is end-of-life, so its packages now live on archive.debian.org.
RUN sed -i -e 's|deb.debian.org|archive.debian.org|' \
        -e 's|security.debian.org|archive.debian.org|' \
        -e '/stretch-updates/d' /etc/apt/sources.list && \
    apt-get -o Acquire::Check-Valid-Until=false update && \
    apt-get install -y --no-install-recommends \
        calibre fonts-dejavu git libnss3 python xdg-utils && \
    rm -rf /var/lib/apt/lists/*

# Let Calibre's ebook-convert run without a display.
ENV QT_QPA_PLATFORM=offscreen

ENV GITBOOK_DIR=/usr/local/lib/gitbook

RUN npm install gitbook-cli -g && \
    gitbook fetch 3.2.3

RUN mkdir /build && chown node:node /build && chmod 0750 /build

USER node

WORKDIR /build
