package relay

import (
	"fmt"
	"sync"
)

// MessageQueue with messages and a fixed size
type MessageQueue struct {
	messageMutex sync.Mutex
	messages     [][]byte
	size         int
}

func (m *MessageQueue) Init(size int) {
	m.messages = make([][]byte, 0)
	m.size = size
}

func (m *MessageQueue) Enqueue(messageIn []byte) error {
	if len(m.messages) == m.size {
		return fmt.Errorf("queue is full")
		return nil
	}
	m.messageMutex.Lock()
	m.messages = append(m.messages, messageIn)
	m.messageMutex.Unlock()
	fmt.Println("Enqueued:", messageIn)
	return nil
}

func (m *MessageQueue) Dequeue() ([]byte, error) {
	if len(m.messages) == 0 {
		return nil, fmt.Errorf("empty queue")
	}
	messageOut := m.messages[0]
	if len(m.messages) == 1 {
		m.messageMutex.Lock()
		m.messages = make([][]byte, 0)
		m.messageMutex.Unlock()
		return messageOut, nil
	}
	m.messageMutex.Lock()
	m.messages = m.messages[1:]
	m.messageMutex.Unlock()
	return messageOut, nil
}
