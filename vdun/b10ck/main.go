package main

import (
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "HTTP attack defense system"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level, l",
			Value: "info",
			Usage: "Log level (options: debug, info, warn, error, fatal)",
		},
	}

	app.Before = func(c *cli.Context) error {
		log.SetOutput(os.Stderr)
		level, err := log.ParseLevel(c.String("log-level"))
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.SetLevel(level)
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:   "master",
			Usage:  "Run master",
			Action: master,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "addr",
					Value: "127.0.0.1:3868",
					Usage: "Listening interface (intranet only!)",
				},
			},
		},
		{
			Name:   "agent",
			Usage:  "Run agent",
			Action: agent,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "addr",
					Value: "127.0.0.1:3868",
					Usage: "Listening interface (intranet only!)",
				},
				cli.StringFlag{
					Name:  "master",
					Usage: "Master's host:port, run agent as a client",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
