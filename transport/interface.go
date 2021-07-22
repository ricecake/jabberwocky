package transport

import (
	"time"

	"github.com/ricecake/karma_chameleon/util"
)

func NewMessage(msgType, msgSubType string, content interface{}) Message {
	return Message{
		Id:      util.CompactUUID(),
		Time:    time.Now(),
		Type:    msgType,
		SubType: msgSubType,
		Content: content,
	}
}
