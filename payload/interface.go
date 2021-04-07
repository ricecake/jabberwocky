package payload

import (
	"context"

	"jabberwocky/storage"
	"jabberwocky/transport"

	"github.com/apex/log"
	"github.com/mitchellh/mapstructure"
)

const script = `
print("Starting");
tail("foobar", function(input) {
        print("GOT: ", input);
});

tail("baz", function(input) {
        print("Other: ", input);
});

print("finished");
`

func Execute(ctx context.Context, msg transport.Message, output chan transport.Message) {
	log.Infof("%+v\n", msg)
	msg.Seq++
	payloadCtx, cancel := context.WithCancel(ctx)

	switch msg.Type {
	case "server":
		var serv storage.Server
		mapstructure.Decode(msg.Content, &serv)
		storage.SaveServer(ctx, serv)
		output <- transport.Message{Type: "reconnect"}
	case "serverList":
		var servs []storage.Server
		mapstructure.Decode(msg.Content, &servs)
		for _, serv := range servs {
			storage.SaveServer(ctx, serv)
		}
	case "script":
		go func() {
			runScript(payloadCtx, script, output)
			cancel()
		}()
	}
}

//todo: a helper function that will calculate if we need to do a reconnection loop.
