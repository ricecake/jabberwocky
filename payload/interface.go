package payload

import (
	"context"

	"jabberwocky/transport"

	"github.com/apex/log"
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
	case "script":
		go func() {
			runScript(payloadCtx, script, output)
			cancel()
		}()
	}
}
