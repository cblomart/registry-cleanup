package main

import (
	"fmt"
	"os"

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

		//
		// repo args
		//

		cli.StringFlag{
			Name:   "repo.fullname",
			Usage:  "repository full name",
			EnvVar: "DRONE_REPO",
		},
		cli.StringFlag{
			Name:   "repo.owner",
			Usage:  "repository owner",
			EnvVar: "DRONE_REPO_OWNER",
		},

		//
		// Parameters
		//
		cli.StringFlag{
			Name:   "config.username",
			Usage:  "Docker username",
			EnvVar: "PLUGIN_USERNAME,DRONE_REPO_OWNER",
		},
		cli.StringFlag{
			Name:   "config.password",
			Usage:  "Docker password",
			EnvVar: "PLUGIN_PASSWORD",
		},
		cli.StringFlag{
			Name:   "config.repo",
			Usage:  "Repository to target",
			EnvVar: "PLUGIN_REPO,DRONE_REPO",
		},
		cli.StringFlag{
			Name:   "config.registry",
			Usage:  "Registry to target",
			EnvVar: "PLUGIN_REGISTRY",
		},
		cli.StringFlag{
			Name:   "config.regex",
			Usage:  "Clean Tags that match regex",
			EnvVar: "PLUGIN_REGEX",
		},
		cli.IntFlag{
			Name:   "config.min",
			Usage:  "Minimum number of tags/images to keep",
			EnvVar: "PLUGIN_MIN",
		},
		cli.DurationFlag{
			Name:   "config.max",
			Usage:  "Maximum age of tags/images",
			EnvVar: "PLUGIN_MAX",
		},
	}

	app.Run(os.Args)
}

func run(c *cli.Context) {
	plugin := Plugin{
		Repo: Repo{
			FullName: c.String("repo.fullname"),
			Owner:    c.String("repo.owner"),
		},
		Config: Config{
			Username: c.String("config.username"),
			Password: c.String("config.password"),
			Repo:     c.String("config.repo"),
			Registry: c.String("config.registry"),
			Regex:    c.String("config.regex"),
			Min:      c.Int("config.min"),
			Max:      c.Duration("config.max"),
		},
	}

	if err := plugin.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
