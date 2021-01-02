To run this, run the server first -- you can either run it locally in Go (you'll need a newer version of Go) via:

```
  cd server
  go build .
  ./chatServer
```

Or you can build the docker image:

```
  cd server
  docker build . -t ReReChat
  docker run -p 8080:8080 ReReChat
```

To run the client, you'll need to build it locally (I haven't Dockerized this yet), so again you'll need a recent version of Node.js (v14 or later) as well as the [Yarn package manager](https://yarnpkg.com/) (it probably would work without Yarn, but I didn't test this with NPM):

```
  cd client
  yarn
  yarn start
```

Open a web browser to http://localhost:3000 and try typing something in the input box (return to send, or click the send button). For more fun, open two tabs and see that they both get different (fake) usernames and see each other's messages!

There are still some bugs, and many feature improvements (notably a *real* authentication system!), but this is the skeleton of the system.
