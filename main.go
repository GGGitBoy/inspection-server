package main

import (
	"fmt"
	"inspection-server/pkg/agent"
	"inspection-server/pkg/db"
	"inspection-server/pkg/schedule"
	"inspection-server/pkg/server"
	"inspection-server/pkg/template"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	VERSION = "dev"
	port    int
)

func main() {
	app := cli.NewApp()
	app.Name = "inspection"
	app.Version = VERSION
	app.Usage = "Provides one-touch inspection capability for clusters."
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "port",
			EnvVar:      "HTTP_PORT",
			Value:       8080,
			Usage:       "The inspection listen port.",
			Destination: &port,
		},
		cli.StringFlag{
			Name:   "serverUrl",
			EnvVar: "SERVER_URL",
			Usage:  "The server url of rancher.",
		},
		cli.StringFlag{
			Name:   "bearerToken",
			EnvVar: "BEARER_TOKEN",
			Usage:  "The bearer token of rancher.",
		},
		cli.BoolFlag{
			Name:   "debug",
			EnvVar: "LOG_LEVEL",
			Usage:  "Set log level to debug.",
		},
	}

	app.Action = func(ctx *cli.Context) error {
		router := server.Start()
		logrus.Infof("server running, listening at: %d\n", port)

		return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	}
	app.Before = before

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func before(ctx *cli.Context) error {
	if ctx.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	err := db.Register()
	if err != nil {
		return err
	}

	err = schedule.Register()
	if err != nil {
		return err
	}

	err = template.Register()
	if err != nil {
		return err
	}

	err = agent.Register()
	if err != nil {
		return err
	}

	return nil
}
