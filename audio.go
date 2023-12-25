package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/wav"
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
		pythonCmd := fmt.Sprintf("python -c \"from gradio_client import Client; client = Client('http://127.0.0.1:7860', verbose=False); result = client.predict('%s', fn_index=0); print(result)\"", text)
		cmd := exec.Command("sh", "-c", pythonCmd)
		cmd.Stdout = &stdoutBuffer
		cmd.Stderr = os.Stderr

		// Run the command
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error running Python command:", err)
		}

		resultText := stdoutBuffer.String()
    	GoogleSpeak(resultText, "en")
		client.Say(channel, fmt.Sprintf("/me %s \n", resultText))
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
		audioData, err := readFile("out.wav")
		if err != nil {
			fmt.Println(err)
			return
		}
		err = conn.WriteMessage(websocket.BinaryMessage, audioData)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func readFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func GoogleSpeak(text, language string) error {
	url := fmt.Sprintf("http://translate.google.com/translate_tts?ie=UTF-8&total=1&idx=0&textlen=32&client=tw-ob&q=%s&tl=%s", url.QueryEscape(text), language)
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer response.Body.Close()

	streamer, format, err := mp3.Decode(response.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer streamer.Close()

	file, err := os.Create("out.wav")
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer file.Close()

	wav.Encode(file, streamer, format)

	return nil
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