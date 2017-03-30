## 3 kinds of chats
Here are 3 approaches to implement a simple web chat on Go:
* [http chat](https://github.com/ss-dev/go-chat-http)
* [comet chat](https://github.com/ss-dev/go-chat-comet)
* [websocket chat](https://github.com/ss-dev/go-chat-websocket)

## WebSocket Chat
This is example of implementation a simple [WebSocket](https://en.wikipedia.org/wiki/WebSocket) chat.
The application uses [Gorilla WebSocket](https://github.com/gorilla/websocket) package.

## Running the example
For running the example you should have working [Go development environment](https://golang.org/doc/install).
You can start the server using the following commands:

    $ go get github.com/ss-dev/go-chat-websocket
    $ cd `go list -f '{{.Dir}}' github.com/ss-dev/go-chat-websocket`
    $ go run app.go

To use the example, open http://localhost:8080/ in your browser.
