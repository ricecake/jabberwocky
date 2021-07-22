package payload

import (
	"context"

	"jabberwocky/storage"
	"jabberwocky/transport"
	"jabberwocky/util"

	"github.com/apex/log"
	"github.com/mitchellh/mapstructure"
)

const script = `
print("Starting");
tail("foobar", function(input) {
        print("GOT: " + input);
});

tail("baz", function(input) {
        print("Other: " + input);
});

print("finished");
`

func Execute(ctx context.Context, msg transport.Message, output chan transport.Message) {
	payloadCtx, cancel := context.WithCancel(ctx)

	switch msg.Type {
	case "server":
		switch msg.SubType {
		case "list":
			var servs []storage.Server
			mapstructure.Decode(msg.Content, &servs)
			err := storage.SaveServers(ctx, servs)
			if err != nil {
				log.Error(err.Error())
			}
			maybeReconnect(ctx, output)
		case "join":
			var serv storage.Server
			mapstructure.Decode(msg.Content, &serv)
			err := storage.SaveServer(ctx, serv)
			if err != nil {
				log.Error(err.Error())
			}
			maybeReconnect(ctx, output)
		case "leave":
			var serv storage.Server
			mapstructure.Decode(msg.Content, &serv)
			err := storage.SaveServer(ctx, serv)
			if err != nil {
				log.Error(err.Error())
			}
			maybeReconnect(ctx, output)
		}
	case "script":
		go func() {
			runScript(payloadCtx, script, output)
			cancel()
		}()
	}
}

func maybeReconnect(ctx context.Context, output chan transport.Message) {
	hrw := util.NewHrw()

	agentId, err := storage.GetNodeId(ctx)
	if err != nil {
		log.Error(err.Error())
	}

	currentServer, err := storage.GetCurrentServer(ctx)
	if err != nil {
		log.Error(err.Error())
	}

	nodes, err := storage.ListLiveServers(ctx)
	if err != nil {
		log.Error(err.Error())
	}

	for _, node := range nodes {
		hrw.AddNode(node)
	}

	newNode := hrw.Get(agentId).(storage.Server)
	if newNode.Uuid != currentServer.Uuid {
		output <- transport.Message{Type: "control", SubType: "reconnect"}
	}
}
