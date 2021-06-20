To run this, run the server first -- you can either run it locally in Go (you'll need a newer version of Go) via:

```
  cd server
  go build .
  ./go_chat
```

Or if you don't have the Golang environment installed (or just want an easier build) you can build the docker image:

```
  cd server
  docker build . -t chat_server:latest
  docker run --rm -it -p 8080:8080 chat_server:latest
```

To run the client, you'll need to build it locally (I haven't Dockerized this yet), so again you'll need a recent version of Node.js (v14 or later) as well as the [Yarn package manager](https://yarnpkg.com/) (it probably would work without Yarn, but I didn't test this with NPM):

```
  cd client
  yarn
  yarn start
```

Open a web browser to http://localhost:3000 to view the client app. Right now, each user to this page is given a different randomly generated name (starting with "tom"). Additionally, a "channel" (`General`) is precreated by the server -- click on it to "join" the channel. This makes it easy to open multiple tabs to the client page and see how the two clients can see each other's actions.

Many bugs remain to be resolved, features to be added...
