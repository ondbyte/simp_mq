package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/ondbyte/simp_mq/simp_broker"
	"github.com/urfave/cli/v2"
)

func main() {
	var broker *simp_broker.SimpBroker
	id, port, token, bufferSize, authWait := "demo_simp_broker", uint(8081), "password", uint(2048), time.Duration(time.Second*10)
	app := &cli.App{
		Name: "simp_mq",
		After: func(ctx *cli.Context) error {
			if broker == nil {
				os.Exit(0)
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "startbroker",
				Usage: "start a SimpBroker",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "name",
						Value:       id,
						Usage:       "name of the SimpBroker instance",
						Destination: &id,
						Aliases:     []string{"n"},
					},
					&cli.UintFlag{
						Name:        "port",
						Value:       port,
						Usage:       "set port for SimpBroker to run on",
						Destination: &port,
						Aliases:     []string{"p"},
					},
					&cli.StringFlag{
						Name:        "token",
						Value:       token,
						Usage:       "SimpBroker will authenticate a client using this token",
						Destination: &token,
						Aliases:     []string{"t"},
					},
					&cli.UintFlag{
						Name:        "buffersize",
						Value:       bufferSize,
						Usage:       "maximum size of the each message exchanged between client and broker",
						Destination: &bufferSize,
						Aliases:     []string{"b"},
					},
					&cli.DurationFlag{
						Name:        "authwait",
						Value:       authWait,
						Usage:       "SimpBroker will wait these milliseconds for authentication from a new connection, after which connection will be dropped",
						Destination: &authWait,
						Aliases:     []string{"w"},
					},
				},
				After: func(ctx *cli.Context) error {
					if broker == nil {
						os.Exit(0)
					}
					return nil
				},
				Action: func(ctx *cli.Context) error {
					broker = &simp_broker.SimpBroker{
						Id:   id,
						Port: fmt.Sprint(port),
						Authenticator: func(deets *simp_broker.AuthDetails) error {
							if deets.Token == token {
								return nil
							}
							return errors.New("failed to authenticate")
						},
						MaxMessageBuffer:          bufferSize,
						DropNoAuthConnectionAfter: time.Duration(1000000 * authWait),
					}

					err := broker.Serve()
					return err
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	close(ch)
	broker.Close()
}
