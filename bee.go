package main

import (
	"os"

	"github.com/urfave/cli"
)

var commands = cli.Commands{}

func registerCommand(c cli.Command) {
	commands = append(commands, c)
}
func main() {
	app := cli.NewApp()
	app.Commands = commands
	app.Name = "Bee"
	app.Version = "0.0.3"
	app.Run(os.Args)
}
