package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	courier "github.com/trisacrypto/courier/pkg"
	"github.com/trisacrypto/courier/pkg/api/v1"
	"github.com/trisacrypto/courier/pkg/config"
	"github.com/trisacrypto/courier/pkg/secrets"
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
			{
				Name:     "store:password",
				Usage:    "store a pkcs12 password using the courier server",
				Category: "client",
				Action:   storePassword,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "url",
						Aliases:  []string{"u", "endpoint"},
						Usage:    "url to connect to the courier server",
						EnvVars:  []string{"COURIER_CLIENT_URL"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "id",
						Aliases:  []string{"i"},
						Usage:    "the id of the certificate to store the password for",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "password",
						Aliases: []string{"p"},
						Usage:   "the password to store",
					},
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "specify a file to read the password from",
					},
				},
			},
			{
				Name:     "store:certificate",
				Usage:    "store a pkcs12 certificate using the courier server",
				Category: "client",
				Action:   storeCertificate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "url",
						Aliases:  []string{"u", "endpoint"},
						Usage:    "url to connect to the courier server",
						EnvVars:  []string{"COURIER_CLIENT_URL"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "id",
						Aliases:  []string{"i"},
						Usage:    "the id of the certificate, used to lookup the stored password",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "path to the certificate file",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "no-decrypt",
						Aliases: []string{"D"},
						Usage:   "do not decrypt the certificate before storing it",
					},
				},
			},
			{
				Name:     "secrets:get",
				Usage:    "get a secret from the secret manager",
				Category: "secrets",
				Action:   getSecret,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project",
						Aliases:  []string{"p"},
						Usage:    "project name where the secret is stored",
						EnvVars:  []string{"COURIER_SECRET_MANAGER_PROJECT"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "name of the secret to get",
						EnvVars:  []string{"COURIER_SECRET_NAME"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    "credentials",
						Aliases: []string{"c"},
						Usage:   "path to the credentials file for the secret manager",
						EnvVars: []string{"GOOGLE_APPLICATION_CREDENTIALS"},
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

// Store a password using the courier service.
func storePassword(c *cli.Context) (err error) {
	var client api.CourierClient
	if client, err = api.New(c.String("url")); err != nil {
		return cli.Exit(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if c.String("password") == "" && c.String("file") == "" {
		return cli.Exit("either --password or --file must be specified", 1)
	}

	var password string
	if password = c.String("password"); password == "" {
		var f *os.File
		if f, err = os.Open(c.String("file")); err != nil {
			return cli.Exit(err, 1)
		}

		var data []byte
		if data, err = io.ReadAll(f); err != nil {
			return cli.Exit(err, 1)
		}

		password = string(data)
	}

	req := &api.StorePasswordRequest{
		ID:       c.String("id"),
		Password: password,
	}
	if err = client.StoreCertificatePassword(ctx, req); err != nil {
		return cli.Exit(err, 1)
	}

	return nil
}

// Store a certificate using the courier service.
func storeCertificate(c *cli.Context) (err error) {
	var client api.CourierClient
	if client, err = api.New(c.String("url")); err != nil {
		return cli.Exit(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var f *os.File
	if f, err = os.Open(c.String("file")); err != nil {
		return cli.Exit(err, 1)
	}

	var data []byte
	if data, err = io.ReadAll(f); err != nil {
		return cli.Exit(err, 1)
	}

	req := &api.StoreCertificateRequest{
		ID:                c.String("id"),
		NoDecrypt:         c.Bool("no-decrypt"),
		Base64Certificate: base64.StdEncoding.EncodeToString(data),
	}
	if err = client.StoreCertificate(ctx, req); err != nil {
		return cli.Exit(err, 1)
	}

	return nil
}

// Get a secret from the secret manager.
func getSecret(c *cli.Context) (err error) {
	conf := config.GCPSecretsConfig{
		Enabled:     true,
		Project:     c.String("project"),
		Credentials: c.String("credentials"),
	}

	secrets, err := secrets.NewClient(conf)
	if err != nil {
		return cli.Exit(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var secret []byte
	if secret, err = secrets.GetLatestVersion(ctx, c.String("name")); err != nil {
		return cli.Exit(err, 1)
	}

	fmt.Println(string(secret))
	return nil
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
