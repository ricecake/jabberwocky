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

/*
TODO:
Instead of an interface, just make a command struct.
Then can define each command more simply, and into variables.
There should be an Init function on the file that will "register" each of the commands, a la cobra,
and from there let the actual engine do it's job of managing commands and executing them.
*/
type command struct {
	name string   // name of function
	path []string // where it lives.  can be empty
	// autocomplete string --- this might be helpful for defining what we want it to autocomplete as
	body    exFunc // the actual function body
	cleanup func() // takes no args, returns no values.  Need to figure out how to pass in the command best. maybe cleanup and body just need to have the command passed in?
	// cleanup func(command)
	// body func(command) exFunc
	ctx     context.Context
	output  chan transport.Message
	runtime *runtime
	state   interface{} // this is whatever the command needs to either work, or cleanup.
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

// TODO: Need to find a way for scripts to include other scripts.  There should be a control where they can only import scripts and functions
// with a equal or lesser access level than the running script.

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

// TODO: Should make all of the commands be more like how cobra commands work, since we don't need multiples of each of the commands.
