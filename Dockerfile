FROM node:12.11.1-slim

RUN apt-get update
RUN apt-get install -y python xdg-utils wget xz-utils git libnss3
RUN wget -nv -O- http://download.calibre-ebook.com/linux-installer.sh | sh /dev/stdin
RUN npm update -g

ENV GITBOOK_DIR=/usr/local/lib/gitbook

RUN npm install gitbook-cli -g && \
    gitbook fetch 3.2.3

RUN mkdir /build && chown node:node /build && chmod 0750 /build

USER node

WORKDIR /build
