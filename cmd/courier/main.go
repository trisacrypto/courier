package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	courier "github.com/trisacrypto/courier/pkg"
	"github.com/trisacrypto/courier/pkg/api/v1"
	"github.com/trisacrypto/courier/pkg/config"
	"github.com/urfave/cli/v2"
)

func main() {
	// Load the dotenv file if it exists
	godotenv.Load()

	// Create the CLI application
	app := &cli.App{
		Name:    "courier",
		Version: courier.Version(),
		Usage:   "a standalone certificate delivery service",
		Flags:   []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:     "serve",
				Usage:    "run the courier server",
				Category: "server",
				Action:   serve,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "addr",
						Aliases:  []string{"a"},
						Usage:    "address:port to bind the server on",
						EnvVars:  []string{"COURIER_BIND_ADDR"},
						Required: true,
					},
				},
			},
			{
				Name:     "status",
				Usage:    "get the status of the courier server",
				Category: "client",
				Action:   status,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "url",
						Aliases:  []string{"u", "endpoint"},
						Usage:    "url to connect to the courier server",
						EnvVars:  []string{"COURIER_CLIENT_URL"},
						Required: true,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

//===========================================================================
// CLI Actions
//===========================================================================

// Serve the courier service.
func serve(c *cli.Context) (err error) {
	var conf config.Config
	if conf, err = config.New(); err != nil {
		return cli.Exit(err, 1)
	}

	if addr := c.String("addr"); addr != "" {
		conf.BindAddr = addr
	}

	var srv *courier.Server
	if srv, err = courier.New(conf); err != nil {
		return cli.Exit(err, 1)
	}

	if err = srv.Serve(); err != nil {
		return cli.Exit(err, 1)
	}

	return nil
}

// Get the status of the courier service.
func status(c *cli.Context) (err error) {
	var client api.CourierClient
	if client, err = api.New(c.String("url")); err != nil {
		return cli.Exit(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rep *api.StatusReply
	if rep, err = client.Status(ctx); err != nil {
		return cli.Exit(err, 1)
	}

	return printJSON(rep)
}

//===========================================================================
// Helpers
//===========================================================================

// Print an object as encoded JSON to stdout.
func printJSON(v interface{}) (err error) {
	var data []byte
	if data, err = json.MarshalIndent(v, "", "  "); err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}
