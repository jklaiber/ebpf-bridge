package service

type Handler interface {
	HandleMessages() error
}

type MessageHandler struct {
	messages  <-chan string
	responses chan<- string
}

func NewMessageHandler(messages <-chan string, responses chan<- string) *MessageHandler {
	return &MessageHandler{messages: messages, responses: responses}
}

func (h *MessageHandler) HandleMessages() error {
	for {
		message := <-h.messages
		h.responses <- message
	}
}
