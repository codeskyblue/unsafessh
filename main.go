package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "unsafessh"
	app.Version = "0.1"
	app.Author = "codeskyblue@gmail.com"
	app.Commands = []cli.Command{
		execCommand,
		servCommand,
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "addr",
			Value: "/tmp/unsafessh.sock",
			Usage: "listen address",
		},
		cli.StringFlag{
			Name:  "proto",
			Value: "unix",
			Usage: "<unix|tcp>",
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
