FROM node:14-alpine

RUN mkdir -p /src
WORKDIR  /src
COPY package.json .
RUN npm install

COPY server.js .
EXPOSE 8080

ENTRYPOINT node server.js