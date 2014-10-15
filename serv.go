package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/gobuild/log"
)

var servCommand = cli.Command{
	Name:   "serv",
	Usage:  "run as server",
	Action: servAction,
	Flags: []cli.Flag{
		cli.BoolFlag{Name: "d,debug"},
	},
}

func servAction(ctx *cli.Context) {
	lis, err := net.Listen(ctx.GlobalString("proto"), ctx.GlobalString("addr"))
	if err != nil {
		log.Fatal(err)
	}
	if ctx.Bool("debug") {
		log.SetOutputLevel(log.Ldebug)
	}

	TrapSignal(func(sig os.Signal) {
		log.Infof("catch signal: %v", sig)
		lis.Close()
		os.Exit(0)
	}, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	for {
		fd, err := lis.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleServer(fd)
	}
}

func handleServer(c net.Conn) error {
	defer c.Close()

	rdch := NewJsonStream(c)
	send := NewJsonSender(c)

	v := <-rdch
	args := make([]string, 0, 10)
	if err := json.Unmarshal([]byte(v.Data), &args); err != nil {
		return err
	}

	log.Infof("COMMAND ARGS: %v", args)
	cmd := exec.Command(args[0], args[1:]...)
	stdinPipe, _ := cmd.StdinPipe()
	cmd.Stderr = NewFuncWriter(func(data string) error {
		send("STDERR", data)
		return nil
	})
	cmd.Stdout = NewFuncWriter(func(data string) error {
		send("STDOUT", data)
		return nil
	})
	if err := cmd.Start(); err != nil {
		send("ERROR", err.Error())
		return err
	}

	exit := Go(cmd.Wait)

	go func() {
		for v := range rdch {
			switch v.Name {
			case "STDIN":
				if v.Data == "" { // EOF
					stdinPipe.Close()
					break
				}
				stdinPipe.Write([]byte(v.Data))
				log.Debugf("STDIN: %#v", v.Data)
			case "SIGNAL":
				var sig syscall.Signal
				fmt.Sscanf(v.Data, "%d", &sig)
				if cmd.Process != nil {
					cmd.Process.Signal(sig)
				}
				break
			default:
				send("ERROR", "unknown command type: "+v.Name)
			}
		}
	}()
	exitCode := ReadExitcode(<-exit)
	log.Infof("DONE, exit=%d", exitCode)
	return send("EXIT", exitCode)
}
