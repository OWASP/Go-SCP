FROM node:current-slim

RUN apt-get update && apt-get install -y calibre

WORKDIR /build
