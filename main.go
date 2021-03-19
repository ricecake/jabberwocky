package main

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/hpcloud/tail"
)

const script = `
print("Starting");
tailLogFile("foobar", function(input) {
	print("GOT: ", input);
});

tailLogFile("baz", function(input) {
	print("Other: ", input);
});

print("finished");
`

/* TODO:
Need sqlite for db
need websocket hookup
need server side
need signing
need more primitive ops
need to make primitive ops more generic to make extensible

need fancy lb stuff on the server side
need to store client jobs in db, and handle init on startup
need a basic kv store in the db, and helpers so scripts can access it
need cron functionality
message envelope should handle types like cron, exec, tail, script.  The cron bit should be a property that can be applied to any job that isn't persistant.  Also want "watch" for inotify stuff

Need a type field, and a interval field.
so exec, tail, script, pipe, watch
and intervals of Once, Cron, Fixed, Boot
Once for a long running thing would just mean run it, and then when it stops, let it be over.
Boot would make it install a long running job when the daemon starts
cron would install a standard cron job
fixed would be once every fixed interval, like "30 minutes"

On startup, install crons
then start interval jobs, starting at a random point in each interval, unless a recorded last run is present
then run all boot jobs
*/

type asyncEvent struct {
	Type string
	Data interface{}
}


type primitive struct {
	name string
	body func(args ... goja.Value)
}


func main() {
	vm := goja.New()

	prim := primitive{
		name: "hello",
		body: func(args ...goja.Value) { fmt.Println("hello")},
	}

	var stopped bool
	callbacks := make(chan func(), 1)
	quit := make(chan int, 1)

	vm.Set(prim.name, prim.body)

	vm.Set("quit", func(data ...goja.Value) {
		if !stopped {
			stopped = true
			close(quit)
		}
	})
	vm.Set("print", func(data ...goja.Value) {
		fmt.Println(data)
	})

	var tails []*tail.Tail

	vm.Set("tailLogFile", func(file, cb goja.Value) {
		var callback func(line string)
		vm.ExportTo(cb, &callback)
		t, err := tail.TailFile(file.Export().(string), tail.Config{Follow: true})
		if err != nil {
			return
		}

		tails = append(tails, t)

		go func() {
			for line := range t.Lines {
				callbacks <- func() {
					callback(line.Text)
				}
			}
			fmt.Println("Listenr ended")
		}()
	})

	//	the tail function should register a tail in a thread, and setup an entry in an eventstream that can listen afterwards if there are any.
	//	Need to make sure that the values and data stays in the same routine as the vm

	//v, err := vm.RunScript("foo", "let a = function () { return \"cats\"}; a")
	_, err := vm.RunScript("foo", script)
	if err != nil {
		panic(err)
	}

	cbLoop: for {
		select {
		case <-quit:
			break cbLoop
		case cb := <-callbacks:
			cb()
		}
	}

	for _, tailer := range tails {
		tailer.Stop()
		tailer.Cleanup()
	}

	//	var f func() string
	//	vm.ExportTo(v, &f)

	//	res := f()

	//	fmt.Printf("%+v\n", res)
	//	if num := v.Export().(int64); num != 4 {
	//		panic(num)
	//	}
}
