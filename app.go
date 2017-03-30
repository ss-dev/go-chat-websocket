package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type RawData struct {
	Name string
	Text string
}

type Message struct {
	Id        int
	Timestamp time.Time
	Name      string
	Text      string
}

type Client struct {
	Connection *websocket.Conn
	SendBuffer chan []byte
}

type Chat struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	QMessages  chan RawData
}

var bufCounter int = -1
var messageList []Message = []Message{}
var chat Chat = Chat{
	Register:   make(chan *Client, 10),
	Unregister: make(chan *Client, 10),
	Clients:    make(map[*Client]bool),
	QMessages:  make(chan RawData, 100),
}
var upgrader = websocket.Upgrader{}

func (c Chat) Run() {
	for {
		select {
		case client := <-c.Register:
			c.Clients[client] = true

			var last5 []Message
			size := len(messageList)
			if size >= 5 {
				last5 = messageList[size-5 : size]
			} else {
				last5 = messageList
			}

			for _, msg := range last5 {
				jsonMsg, err := json.Marshal(msg)
				if err == nil {
					client.SendBuffer <- jsonMsg
				}
			}
		case client := <-c.Unregister:
			if _, ok := c.Clients[client]; ok {
				delete(c.Clients, client)
				close(client.SendBuffer)
			}
		case data := <-c.QMessages:
			bufCounter++
			msg := Message{
				Id:        bufCounter,
				Timestamp: time.Now(),
				Name:      data.Name,
				Text:      data.Text,
			}
			messageList = append(messageList, msg)

			jsonMsg, err := json.Marshal(msg)
			if err == nil {
				for client := range c.Clients {
					select {
					case client.SendBuffer <- jsonMsg:
					default:
						c.Unregister <- client
					}
				}
			}
		}
	}
}

func chatServer(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	client := &Client{
		Connection: conn,
		SendBuffer: make(chan []byte, 10),
	}
	chat.Register <- client
	defer func() {
		chat.Unregister <- client
	}()

	go func() {
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("read error:", mt, msg, err)
				return
			}
			var rawMsg RawData
			err = json.Unmarshal(msg, &rawMsg)
			if err != nil {
				fmt.Println("read error:", mt, msg, err)
				return
			}
			chat.QMessages <- rawMsg
		}
	}()

	for {
		msg := <-client.SendBuffer
		err = conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
	}
}

func homeHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadFile("static/index.html")
	if err == nil {
		fmt.Fprintf(w, string(data))
	} else {
		log.Println(err)
		fmt.Fprintf(w, "<html><head></head><body><h1>Error on reading index.html</h1></body></html>")
	}
}

func main() {
	go chat.Run()

	fs := http.FileServer(http.Dir("static/js"))
	http.Handle("/js/", http.StripPrefix("/js/", fs))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/chat/", chatServer)

	var appPort = flag.String("port", "8080", "HTTP service port")
	flag.Parse()

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":"+*appPort, nil))
}
