package main

import (
  	"bytes"
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
		var stdoutBuffer bytes.Buffer
    	pythonCmd := fmt.Sprintf("python -c \"import replicate; print(replicate.run('playgroundai/playground-v2-1024px-aesthetic:42fe626e41cc811eaf02c94b892774839268ce1994ea778eba97103fe1ef51b8', input={'prompt': '%s', 'num_inference_steps': 25})[0])\"", text)
		// pythonCmd := fmt.Sprintf("python -c \"import replicate; print(replicate.run('lucataco/dreamshaper-xl-turbo:0a1710e0187b01a255302738ca0158ff02a22f4638679533e111082f9dd1b615', input={'prompt': '%s', 'num_inference_steps': 5})[0])\"", text)
		// pythonCmd := fmt.Sprintf("python -c \"import replicate; print(replicate.run('lucataco/sdxl-lcm:fbbd475b1084de80c47c35bfe4ae64b964294aa7e237e6537eed938cfd24903d', input={'prompt': '%s', 'num_inference_steps': 4})[0])\"", text)
		cmd := exec.Command("sh", "-c", pythonCmd)
		cmd.Stdout = &stdoutBuffer
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error running Python command:", err)
		}
		resultText := stdoutBuffer.String()
		messageChannel <- resultText
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
