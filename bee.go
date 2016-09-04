package main

import (
	"os"

	"github.com/saturn4er/bee/lib"
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
	app.Version = "0.0.4"
	app.Before = func(*cli.Context) error {
		lib.LoadConfig()
		return nil
	}
	app.Run(os.Args)
}
