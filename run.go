package main

import (
	"github.com/urfave/cli"
	"os"
	"path"
	"strings"
	"github.com/saturn4er/bee/lib"
)

func init() {
	registerCommand(cli.Command{
		Name:    "run",
		Aliases: []string{"r"},
		Usage:   "Run command will supervise the file system of the beego project using inotify, it will recompile and restart the app after any modifications.",

		HelpName:  "",
		ArgsUsage: "[appname] [watchall] [-main=*.go] [-downdoc=true]  [-gendoc=true]  [-e=Godeps -e=folderToExclude]  [-tags=goBuildTags]",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "watchall"},
		},
		Action: func(c *cli.Context) error {
			var appname string
			cwd, _ := os.Getwd()
			if c.Args().First() == "" {
				appname = path.Base(cwd)
				lib.LogInfo("Uses '%s' as 'appname'", appname)
			} else {
				appname = c.Args().First()
				lib.LogInfo("Uses '%s' as 'appname'", appname)
				if strings.HasSuffix(appname, ".go") && lib.FileExists(path.Join(cwd, appname)) {
					lib.LogWarning("The appname has conflic with crupath's file, do you want to build appname as %s\n", appname)
					lib.LogInfo("Do you want to overwrite it? [yes|no]]  ")
					if !lib.AskForConfirmation() {
						return nil
					}
				}
			}
			beegoApplication, err := lib.NewApplication(appname, cwd)
			if err != nil {
				return err
			}
			beegoApplication.Build()
			beegoApplication.Run()
			beegoApplication.RunRestartWatcher()
			<-beegoApplication.ExitC
			return nil
		},
	})
}
