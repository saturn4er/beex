package main

import (
	"github.com/urfave/cli"
	"os"
)

var commands = cli.Commands{}

func registerCommand(c cli.Command) {
	commands = append(commands, c)
}
func main() {
	app := cli.NewApp()
	app.Commands = commands
	app.Name = "Bee"
	app.Version = "0.0.2"
	app.Run(os.Args)
}
