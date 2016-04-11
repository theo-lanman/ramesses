package message

import (
	"testing"

	"encoding/binary"
)

// IdBytes should return a big-endian byte slice representation of the
// Message's id.
func TestIdBytes(t *testing.T) {
	id := uint64(3)

	msg := NewMessage(id, []byte{})

	if id == binary.BigEndian.Uint64(msg.IdBytes()) {
		return
	}

	t.Fail()
}
