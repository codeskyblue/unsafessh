package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/gobuild/log"
)

var execCommand = cli.Command{
	Name:   "exec",
	Usage:  "execute a new command",
	Action: execAction,
	Flags: []cli.Flag{
		cli.BoolFlag{Name: "d,debug"},
	},
}

func execAction(ctx *cli.Context) {
	c, err := net.Dial(ctx.GlobalString("proto"), ctx.GlobalString("addr"))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	rdch := NewJsonStream(c)
	send := NewJsonSender(c)

	if len(ctx.Args()) == 0 {
		log.Fatal("error: args needed")
	}
	log.Debug(ctx.GlobalString("addr"), ctx.Args())
	args, _ := json.Marshal(ctx.Args())
	send("COMMAND", string(args))

	TrapSignal(func(sig os.Signal) {
		if err := send("SIGNAL", fmt.Sprintf("%d", sig)); err != nil {
			log.Fatal(err)
		}
	}, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	go func() {
		for v := range rdch {
			switch v.Name {
			case "EXIT":
				var exitCode int
				fmt.Sscanf(v.Data, "%d", &exitCode)
				os.Exit(exitCode)
			case "STDOUT":
				os.Stdout.WriteString(v.Data)
			case "STDERR":
				os.Stderr.WriteString(v.Data)
			case "ERROR":
				fmt.Println(v.Data)
				os.Exit(1)
			default:
				log.Warn("Unknown data:", v)
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
	fmt.Println("done")
}
