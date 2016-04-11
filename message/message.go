package message

import (
	"encoding/binary"
	"time"
)

// Message models a message.
type Message struct {
	Attempts    int
	CreatedAt   time.Time
	AttemptedAt time.Time
	Id          uint64
	Body        []byte
}

// NewMessage creates a new Message given an id and a body.
func NewMessage(id uint64, body []byte) Message {
	return Message{Attempts: 0, CreatedAt: time.Now(), Id: id, Body: body}
}

// Batch is just a slice of Messages.
type Batch []Message

// NewBatch makes a new Batch preallocated for 20 Messages, because that is the
// current hard-coded max batch size.
func NewBatch() Batch {
	return make([]Message, 0, 20)
}

// IdBytes returns the Message's id as a big-endian slice of bytes, such that
// the resulting byte slice sort order matches the integer sort order.
func (msg Message) IdBytes() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, msg.Id)
	return b
}
