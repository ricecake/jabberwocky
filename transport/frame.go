package transport

import "encoding/json"

type Message struct {
	Seq      int
	Id       string
	Type     string
	Tags     map[string]string
	Metadata interface{}
}

func (m Message) EncodeJson() ([]byte, error) {
	return json.Marshal(m)
}
