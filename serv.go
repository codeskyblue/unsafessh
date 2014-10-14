package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gobuild/log"

	"github.com/codegangsta/cli"
)

var servCommand = cli.Command{
	Name:   "serv",
	Usage:  "run as server",
	Action: servAction,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "addr", Value: "unix:/tmp/unsafessh.sock", Usage: "listen address"},
	},
}

func splitAddr(uri string) (prot string, path string) {
	vs := strings.Split(uri, ":")
	return vs[0], vs[1]
}

func servAction(ctx *cli.Context) {
	prot, value := splitAddr(ctx.String("addr"))
	lis, err := net.Listen(prot, value)
	if err != nil {
		log.Fatal(err)
	}
	defer lis.Close()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		for sig := range sigCh {
			fmt.Printf("received: %v\n", sig)
			lis.Close()
			os.Exit(0)
		}
	}()

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
	rdch := make(chan *Protocol)

	go func() {
		decoder := json.NewDecoder(c)
		for {
			v := new(Protocol)
			err := decoder.Decode(v)
			if err != nil {
				if err == io.EOF {
					log.Println("connection closed")
				}
				log.Println(err)
				break
			}
			rdch <- v
		}
		close(rdch)
	}()

	encoder := json.NewEncoder(c)
	send := func(name, data string) error {
		fmt.Println("Send", name, data)
		return encoder.Encode(&Protocol{Name: name, Data: data})
	}

	v := <-rdch
	fmt.Println(v)
	args := make([]string, 0, 10)
	if err := json.Unmarshal([]byte(v.Data), &args); err != nil {
		return err
	}

	fmt.Println("ARGS:", args)
	cmd := exec.Command(args[0], args[1:]...)
	stdinPipe, _ := cmd.StdinPipe()
	errpipe, _ := cmd.StderrPipe()
	outpipe, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		send("ERROR", err.Error())
		return err
	}

	drainOutput := func(name string, rd io.Reader) {
		buf := make([]byte, 255)
		for {
			nr, err := rd.Read(buf)
			if nr != 0 {
				send(name, string(buf[:nr]))
			}
			if err != nil {
				break
			}
		}
		fmt.Println(name, "DONE")
	}
	go drainOutput("STDERR", errpipe)
	go drainOutput("STDOUT", outpipe)

	exit := Go(cmd.Wait)

	go func() {
		for v := range rdch {
			switch v.Name {
			case "STDIN":
				stdinPipe.Write([]byte(v.Data))
				fmt.Print(v.Data)
			case "SIGNAL":
				var sig syscall.Signal
				fmt.Sscanf(v.Data, "%d", &sig)
				if cmd.Process != nil {
					cmd.Process.Signal(sig)
				}
				break
			default:
				fmt.Println(v)
				send("ERROR", "unknown command type: "+v.Name)
			}
		}
	}()
	err := <-exit
	if err != nil {
		send("EXIT", "1")
	} else {
		send("EXIT", "0")
	}
	return nil
}
