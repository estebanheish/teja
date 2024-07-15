package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"teja/internal/config"
)

type Body struct {
	Model    string
	Messages []Message
}

type Message struct {
	Role    string
	Content string
}

type StreamResponse struct {
	Message Message
	Done    bool
}

func Chat(cfg config.Config, msgs []Message) <-chan string {
	rx := make(chan string)

	systemMessages := make([]Message, len(cfg.Profile.System))
	for i, s := range cfg.Profile.System {
		systemMessages[i] = Message{
			Role:    "system",
			Content: s,
		}
	}

	body, err := json.Marshal(Body{
		Model:    cfg.Profile.Model,
		Messages: append(systemMessages, msgs...),
	})
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(fmt.Sprintf("http://localhost:%s/api/chat", cfg.Ollama.Port), "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		for {
			var message StreamResponse
			if err := decoder.Decode(&message); err == io.EOF {
				break
			} else if err != nil {
				log.Fatalln(err)
			}

			if !message.Done {
				rx <- message.Message.Content
			}
		}
		close(rx)
	}()

	return rx
}
