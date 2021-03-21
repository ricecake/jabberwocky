package payload

import (
	"jabberwocky/transport"

	"github.com/apex/log"
)

func Execute(msg transport.Message, output chan transport.Message) {
	log.Infof("%+v\n", msg)
	msg.Seq++
	output <- msg
}
