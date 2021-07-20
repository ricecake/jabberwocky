package payload

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
	"github.com/hpcloud/tail"
	//	"github.com/apex/log"

	"jabberwocky/transport"
)

type exFunc func(...goja.Value)

type external interface {
	Init(ctx context.Context, rt *runtime, output chan transport.Message)
	Name() string
	Body() exFunc
	Cleanup()
}

type primitive struct {
	name    string
	body    exFunc
	ctx     context.Context
	output  chan transport.Message
	runtime *runtime
}

func (extern *primitive) Init(ctx context.Context, rt *runtime, output chan transport.Message) {
	extern.ctx = ctx
	extern.output = output
	extern.runtime = rt
}

func (ex *primitive) Name() string {
	return ex.name
}

func (ex *primitive) Cleanup() {
	return
}

type externPrint struct {
	primitive
}

func (ex *externPrint) Body() exFunc {
	return func(args ...goja.Value) {
		for _, i := range args {
			//ex.output <- transport.Message{Type: i.Export().(string)}
			ex.output <- transport.Message{
				Type:    "output",
				SubType: "log",
				Content: i.Export().(string),
			}
		}
		fmt.Println(args)
	}
}

type externTail struct {
	primitive
	tails []*tail.Tail
}

func (ex *externTail) Body() exFunc {
	return func(args ...goja.Value) {
		fmt.Println("adding tails")
		file := args[0]
		cb := args[1]

		var callback func(line string)
		ex.runtime.vm.ExportTo(cb, &callback)
		t, err := tail.TailFile(file.Export().(string), tail.Config{Follow: true})
		if err != nil {
			return
		}

		ex.tails = append(ex.tails, t)

		go func() {
			for line := range t.Lines {
				ex.runtime.callbacks <- func() {
					callback(line.Text)
				}
			}
			fmt.Println("Listenr ended")
		}()
	}
}

func (ex *externTail) Cleanup() {
	for _, tailer := range ex.tails {
		tailer.Stop()
		tailer.Cleanup()
	}
}

type externQuit struct {
	primitive
}

func (ex *externQuit) Body() exFunc {
	return func(args ...goja.Value) {
		ex.runtime.shutdown()
	}
}

type externHttpReq struct {
	primitive
}

/*
The body should make the request, and store it in itself, and if canceled, cancel the request.
Needs to call a callback on success, and a different one on error.  So should support
url
args => map with args, headers, body, method and the like
success callback
error callback
*/

type runtime struct {
	vm        *goja.Runtime
	stopped   bool
	callbacks chan func()
	quit      chan struct{}
	externs   []external
}

func (r *runtime) shutdown() {
	if !r.stopped {
		r.stopped = true
		close(r.quit)
	}
}

func (r *runtime) cleanup() {
	for _, ex := range r.externs {
		ex.Cleanup()
	}
}

func createRuntime(ctx context.Context, output chan transport.Message) *runtime {
	runtime := &runtime{
		vm:        goja.New(),
		callbacks: make(chan func()),
		quit:      make(chan struct{}),
		externs: []external{
			&externPrint{primitive{name: "print"}},
			&externTail{primitive: primitive{name: "tail"}},
			&externQuit{primitive{name: "quit"}},
		},
	}

	for _, extern := range runtime.externs {
		extern.Init(ctx, runtime, output)
		runtime.vm.Set(extern.Name(), extern.Body())
	}

	return runtime
}

func runScript(ctx context.Context, script string, output chan transport.Message) {
	rt := createRuntime(ctx, output)
	defer rt.cleanup()

	_, err := rt.vm.RunString(script)
	if err != nil {
		panic(err)
	}

cbLoop:
	for {
		select {
		case <-rt.quit:
			break cbLoop
		case <-ctx.Done():
			rt.shutdown()
			// TODO: also interupt the vm
		case cb := <-rt.callbacks:
			cb()
		}
	}

}
