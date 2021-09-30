package transport

import (
	"encoding/json"
	"time"

	"jabberwocky/storage"

	"github.com/mitchellh/mapstructure"
)

type Message struct {
	Id       string
	SourceId string
	JobId    string
	Seq      int
	Time     time.Time
	Type     string
	SubType  string // might be useful for telling the type output.  One output type, many output subtypes. -- this might be better thought of as "action", and type as "content type"
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

type AgentIdentity struct {
	//	Hostname    string // this should just be a tag, but don't want to forget
	Uuid        string
	PublicKey   string
	PublicKeyId string
	Tags        map[string]string // thses should be pulled from system, as well as merged from config file.  used for payload distribution
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

func DecodeJson(msg []byte) (Message, error) {
	var decoded Message
	err := json.Unmarshal(msg, &decoded)
	if err != nil {
		return decoded, err
	}

	// TODO: make this then decide what the type of content is, by having a map function of message type to content type.
	if contentType, matched := stringToType(decoded.Type, decoded.SubType); matched {
		mapstructure.Decode(decoded.Content, contentType)
		decoded.Content = contentType
	}

	return decoded, err
}

func stringToType(typeName, subTypeName string) (interface{}, bool) {
	switch typeName {
	case "script":
		return &storage.Script{}, true
	case "server":
		if subTypeName == "list" {
			return &[]storage.Server{}, true
		}
		return &storage.Server{}, true
	case "agent":
		return &storage.Agent{}, true
	default:
		return nil, false
	}
}
