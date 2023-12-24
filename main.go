package main

import (
	"fmt"
  	"os"
	"os/exec"
	"io/ioutil"
	"net/http"

	"github.com/gempir/go-twitch-irc/v3"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func taskTwitch(clientUsername, channel, oauthToken string, messageChannel chan<- string) {
	client := twitch.NewClient(clientUsername, oauthToken)
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		text := message.Message
		// client.Say(channel, fmt.Sprintf("/me %s \n", text))
    		pythonCmd := fmt.Sprintf("python -c \"from gradio_client import Client; client = Client('http://127.0.0.1:7860'); result = client.predict('%s', fn_index=0); print(result)\"", text)
		cmd := exec.Command("sh", "-c", pythonCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	
		// Run the command
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error running Python command:", err)
		}

		messageChannel <- text
	})
	client.Join(channel)
	err := client.Connect()
	if err != nil {
		panic(err)
	}
}

func handleWebSocketConnection(conn *websocket.Conn, messageChannel <-chan string) {
	defer conn.Close()
	fmt.Println("Client connected")
	for {
		message := <-messageChannel
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			fmt.Println(err)
			return
		}
		imageData, err := readFile("image.jpg")
		if err != nil {
			fmt.Println(err)
			return
		}
		err = conn.WriteMessage(websocket.BinaryMessage, imageData)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func readFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func main() {
	messageChannel := make(chan string)
	go taskTwitch("username_sed", "channel_sed", "oauth_sed", messageChannel)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		handleWebSocketConnection(conn, messageChannel)
	})
	fmt.Println("Server is running on :80")
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		fmt.Println(err)
	}
}
