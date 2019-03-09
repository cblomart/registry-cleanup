package main

import (
	"fmt"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"
)

var version string // build number set at compile-time

func main() {
	app := cli.NewApp()
	app.Name = "registry-cleanup"
	app.Usage = "Clean a registry repository from lingering tags/images"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "username",
			Usage:  "Docker username",
			EnvVar: "PLUGIN_USERNAME,DRONE_REPO_OWNER",
		},
		cli.StringFlag{
			Name:   "password",
			Usage:  "Docker password",
			EnvVar: "PLUGIN_PASSWORD",
		},
		cli.StringFlag{
			Name:   "repo",
			Usage:  "Repository to target",
			EnvVar: "PLUGIN_REPO,DRONE_REPO",
		},
		cli.StringFlag{
			Name:   "registry",
			Value:  DefaultRegistry,
			Usage:  "Registry to target",
			EnvVar: "PLUGIN_REGISTRY",
		},
		cli.StringFlag{
			Name:   "regex",
			Value:  "^[0-9A-Fa-f]+$",
			Usage:  "Clean Tags that match regex",
			EnvVar: "PLUGIN_REGEX",
		},
		cli.IntFlag{
			Name:   "min",
			Value:  3,
			Usage:  "Minimum number of tags/images to keep",
			EnvVar: "PLUGIN_MIN",
		},
		cli.DurationFlag{
			Name:   "max",
			Value:  360 * time.Hour,
			Usage:  "Maximum age of tags/images",
			EnvVar: "PLUGIN_MAX",
		},
		cli.BoolFlag{
			Name:   "verbose",
			Usage:  "Show verbose information",
			EnvVar: "PLUGIN_VERBOSE",
		},
		cli.BoolFlag{
			Name:   "dryrun",
			Usage:  "Dry run",
			EnvVar: "PLUGIN_DRYRUN",
		}}

	app.Run(os.Args)
}

func run(c *cli.Context) {
	plugin := Plugin{
		Username: c.String("username"),
		Password: c.String("password"),
		Repo:     c.String("repo"),
		Registry: c.String("registry"),
		Regex:    c.String("regex"),
		Min:      c.Int("min"),
		Max:      c.Duration("max"),
		Verbose:  c.Bool("verbose"),
		DryRun:   c.Bool("dryrun"),
	}

	if err := plugin.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
