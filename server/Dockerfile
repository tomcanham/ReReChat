# build stage
FROM golang:1.15-alpine AS build-env
RUN apk --no-cache add build-base

WORKDIR /go/src/app
COPY . .

RUN cd /go/src/app && go build -o server

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /go/src/app/server /app/

CMD ["./server"]