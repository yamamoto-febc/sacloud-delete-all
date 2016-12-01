// sacloud-delete-all : A CLI tool of to delete all resources on Sakura Cloud.
package main

import (
	"fmt"
	"github.com/yamamoto-febc/sacloud-delete-all/lib"
	"github.com/yamamoto-febc/sacloud-delete-all/version"
	"gopkg.in/urfave/cli.v2"
	"io"
	"os"
	"strings"
)

var (
	appName              = "sacloud-delete-all"
	appUsage             = "A CLI tool of to delete all resources on Sakura Cloud"
	appCopyright         = "Copyright (C) 2016 Kazumichi Yamamoto."
	applHelpTextTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.HelpName}} [options]

REQUIRED PARAMETERS:
   {{range .RequiredFlags}}{{.}}
   {{end}}
OPTIONS:
   {{range .OptionFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}

{{.Copyright}}
`

	requiredFlagNames = []string{
		"token",
		"secret",
	}
)

func main() {

	cli.AppHelpTemplate = applHelpTextTemplate
	app := &cli.App{}
	option := lib.NewOption()

	app.Name = appName
	app.Usage = appUsage
	app.HelpName = appName
	app.Copyright = appCopyright

	app.Flags = cliFlags(option)
	app.Action = cliCommand(option)
	app.Version = version.FullVersion()

	originalHelpPrinter := cli.HelpPrinter
	cli.HelpPrinter = func(w io.Writer, templ string, d interface{}) {
		app := d.(*cli.App)
		data := newHelpData(app)
		originalHelpPrinter(w, templ, data)
	}

	app.Run(os.Args)
}

type helpData struct {
	*cli.App
	RequiredFlags []cli.Flag
	OptionFlags   []cli.Flag
}

func newHelpData(app *cli.App) interface{} {
	data := &helpData{App: app}

	for _, f := range app.VisibleFlags() {
		if isExistsFlag(requiredFlagNames, f) {
			data.RequiredFlags = append(data.RequiredFlags, f)
		} else {
			data.OptionFlags = append(data.OptionFlags, f)
		}
	}

	return data
}

func cliFlags(option *lib.Option) []cli.Flag {

	return []cli.Flag{
		&cli.StringFlag{
			Name:        "token",
			Aliases:     []string{"sakuracloud-access-token"},
			Usage:       "API Token of SakuraCloud",
			EnvVars:     []string{"SAKURACLOUD_ACCESS_TOKEN"},
			DefaultText: "none",
			Destination: &option.AccessToken,
		},
		&cli.StringFlag{
			Name:        "secret",
			Aliases:     []string{"sakuracloud-access-token-secret"},
			Usage:       "API Secret of SakuraCloud",
			EnvVars:     []string{"SAKURACLOUD_ACCESS_TOKEN_SECRET"},
			DefaultText: "none",
			Destination: &option.AccessTokenSecret,
		},
		&cli.StringSliceFlag{
			Name:    "zones",
			Aliases: []string{"sakuracloud-zones"},
			Usage:   "Target zone list of SakuraCloud",
			EnvVars: []string{"SAKURACLOUD_ZONES"},
			Value:   cli.NewStringSlice(lib.SakuraCloudDefaultZones...),
		},
		&cli.BoolFlag{
			Name:        "sakuracloud-trace-mode",
			Usage:       "Flag of SakuraCloud debug-mode",
			EnvVars:     []string{"SAKURACLOUD_TRACE_MODE"},
			Destination: &option.TraceMode,
			Value:       false,
		},
		&cli.BoolFlag{
			Name:        "trace-log",
			Usage:       "Flag of enable TRACE log",
			EnvVars:     []string{"SACKEREL_TRACE_LOG"},
			Destination: &option.JobQueueOption.TraceLog,
			Value:       false,
		},
		&cli.BoolFlag{
			Name:        "info-log",
			Usage:       "Flag of enable INFO log",
			EnvVars:     []string{"SACKEREL_INFO_LOG"},
			Value:       true,
			Destination: &option.JobQueueOption.InfoLog,
		},
		&cli.BoolFlag{
			Name:        "warn-log",
			Usage:       "Flag of enable WARN log",
			EnvVars:     []string{"SACKEREL_WARN_LOG"},
			Value:       true,
			Destination: &option.JobQueueOption.WarnLog,
		},
		&cli.BoolFlag{
			Name:        "error-log",
			Usage:       "Flag of enable ERROR log",
			EnvVars:     []string{"SACKEREL_ERROR_LOG"},
			Value:       true,
			Destination: &option.JobQueueOption.ErrorLog,
		},
	}

}

func cliCommand(option *lib.Option) func(c *cli.Context) error {
	return func(c *cli.Context) error {

		option.Zones = c.StringSlice("zones")
		errors := option.Validate()
		if len(errors) != 0 {
			return flattenErrors(errors)
		}

		return lib.Run(option)

	}
}

func flattenErrors(errors []error) error {
	var list = make([]string, 0)
	for _, str := range errors {
		list = append(list, str.Error())
	}
	return fmt.Errorf(strings.Join(list, "\n"))
}

func isExistsFlag(source []string, target cli.Flag) bool {
	for _, s := range source {
		if s == target.Names()[0] {
			return true
		}
	}
	return false
}
