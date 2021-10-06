package cluster

import (
	"sort"

	"github.com/ricecake/karma_chameleon/util"

	"jabberwocky/transport"
)

type clusterEnvelope struct {
	Id      string
	Server  string
	Emitter Emitter
	Message transport.Message
}

func packageMessage(e Emitter, msg transport.Message) clusterEnvelope {
	return clusterEnvelope{
		Id:      util.CompactUUID(),
		Server:  handler.nodeId,
		Emitter: e,
		Message: msg,
	}
}

type Emitter int

const (
	LOCAL_CLIENT Emitter = iota
	LOCAL_AGENT
	LOCAL_SERVER
	PEER_CLIENT
	PEER_AGENT
	PEER_SERVER
)

func (e Emitter) String() string {
	return [...]string{
		"Local Client",
		"Local Agent",
		"Local Server",
		"Peer Client",
		"Peer Agent",
		"Peer Server",
	}[e]
}

func (e Emitter) ConvertToPeer() Emitter {
	return [...]Emitter{
		PEER_CLIENT,
		PEER_AGENT,
		PEER_SERVER,
		PEER_CLIENT,
		PEER_AGENT,
		PEER_SERVER,
	}[e]
}

func (e Emitter) IsLocal() bool {
	return [...]bool{
		true,
		true,
		true,
		false,
		false,
		false,
	}[e]
}

type bindEntry struct {
	key   string
	value string
}
type Destination struct {
	Role Emitter
	Code string
}

func linearizeTags(bind map[string]string) (bindings []bindEntry) {
	for k, v := range bind {
		bindings = append(bindings, bindEntry{key: k, value: v})
	}

	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].key != bindings[j].key {
			return bindings[i].key < bindings[j].key
		}
		return bindings[i].value < bindings[j].value
	})

	return
}

func copyTagMap(original map[string]string) map[string]string {
	copy := make(map[string]string)
	for key, value := range original {
		copy[key] = value
	}
	return copy
}

func deduplicate(destinations []Destination) (dedupe []Destination) {
	return
}
