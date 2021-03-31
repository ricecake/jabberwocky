package transport

import (
	"encoding/json"
	"time"
)

type Message struct {
	Id       string
	JobId    string
	Seq      int
	Time     time.Time
	Type     string
	Tags     map[string]string
	Metadata interface{}
	Content  interface{}
}

type Command struct {
	IntervalType  string // once fixed cron boot connect
	FixedInterval int
	CronInterval  string
	Type          string // exec script watch tail journal, and also things like "cancel a job"
	Payload       string
	Arguments     string
	MaxDuration   int
}

type Challange struct {
	Kid   string
	Alg   string
	Typ   string
	Time  string
	Nonce string
}

type ChallangeResponce struct {
	Challange Challange
	Response  string
}

type SetTags struct {
	Hostname    string
	AgentId     string
	PublicKey   string
	PublicKeyId string
	Tags        map[string]string
}

type SetStatus struct {
	Status string
	Attr   map[string]interface{}
}

type LogOutput struct {
	Level   string
	Time    time.Time
	Attr    map[string]string
	Message string
}

type ExecOutput struct {
	Channel string
	Content string
	Exit    int
	Pid     int
	Time    time.Time
}

func (m Message) EncodeJson() ([]byte, error) {
	return json.Marshal(m)
}
