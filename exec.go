package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gobuild/log"

	"github.com/codegangsta/cli"
)

var execCommand = cli.Command{
	Name:   "exec",
	Usage:  "execute a new command",
	Action: execAction,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "addr", Value: "unix:/tmp/unsafessh.sock", Usage: "listen address"},
	},
}

func execAction(ctx *cli.Context) {
	c, err := net.Dial("unix", "/tmp/unsafessh.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	rdch := make(chan *Protocol)
	go func() {
		decoder := json.NewDecoder(c)
		for {
			v := new(Protocol)
			err := decoder.Decode(v)
			if err != nil {
				rdch <- &Protocol{Name: "ERROR", Data: "deocde error"}
				break
			}
			rdch <- v
		}
	}()

	encoder := json.NewEncoder(c)
	send := func(name, data string) error {
		v := &Protocol{Name: name, Data: data}
		return encoder.Encode(v)
	}

	fmt.Println(ctx.String("addr"), ctx.Args())
	args, _ := json.Marshal(ctx.Args())
	send("COMMAND", string(args))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		for sig := range sigCh {
			send("SIGNAL", fmt.Sprintf("%d", sig))
		}
	}()

	go func() {
		for v := range rdch {
			switch v.Name {
			case "EXIT":
				var exitCode int
				fmt.Sscanf(v.Data, "%d", &exitCode)
				os.Exit(exitCode)
			case "STDOUT":
				fmt.Print(v.Data)
			case "ERROR":
				fmt.Println(v.Data)
				os.Exit(1)
			default:
				fmt.Println(v)
			}
		}
	}()
	buf := make([]byte, 256)
	for {
		nr, err := os.Stdin.Read(buf)
		if err != nil {
			send("SIGNAL", "NOINPUT")
		}
		send("STDIN", string(buf[:nr]))
	}
	c.Write(nil)
	fmt.Println("done")
}
