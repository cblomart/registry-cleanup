//   Copyright (c) 2019 cblomart
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

//go:generate git-version

import (
	"fmt"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "registry-cleanup"
	app.Usage = "Clean a registry repository from lingering tags/images"
	app.Action = run
	app.Version = fmt.Sprintf("%s - %s (%s)", gitTag, gitShortCommit, gitStatus)
	app.Authors = []cli.Author{
		cli.Author
		{
			Name:  "Cédric Blomart",
			Email: "cblomart@gmail.com",
		},
	}
	app.Copyright = "Copyright (c) 2019 cblomart"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "username, u",
			Usage:  "Docker username",
			EnvVar: "PLUGIN_USERNAME,DOCKER_USERNAME,DRONE_REPO_OWNER",
		},
		cli.StringFlag{
			Name:   "password, p",
			Usage:  "Docker password",
			EnvVar: "PLUGIN_PASSWORD,DOCKER_PASSWORD",
		},
		cli.StringFlag{
			Name:   "repo, r",
			Usage:  "Repository to target",
			EnvVar: "PLUGIN_REPO,DRONE_REPO",
		},
		cli.StringFlag{
			Name:   "registry",
			Value:  DefaultRegistry,
			Usage:  "Registry to target",
			EnvVar: "PLUGIN_REGISTRY",
		},
		cli.BoolFlag{
			Name:   "insecure, i",
			Usage:  "Skip TLS verification",
			EnvVar: "PLUGIN_INSECURE",
		},
		cli.StringFlag{
			Name:   "regex",
			Value:  "^[0-9A-Fa-f]+$",
			Usage:  "Clean Tags that match regex",
			EnvVar: "PLUGIN_REGEX",
		},
		cli.IntFlag{
			Name:   "min, m",
			Value:  3,
			Usage:  "Minimum number of tags/images to keep",
			EnvVar: "PLUGIN_MIN",
		},
		cli.DurationFlag{
			Name:   "max, M",
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
		},
		cli.BoolFlag{
			Name:   "dump",
			Usage:  "Dump network requests",
			EnvVar: "PLUGIN_DUMP",
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
	}
}

func run(c *cli.Context) {
	plugin := Plugin{
		Username: c.String("username"),
		Password: c.String("password"),
		Repo:     c.String("repo"),
		Registry: c.String("registry"),
		Insecure: c.Bool("insecure"),
		Regex:    c.String("regex"),
		Min:      c.Int("min"),
		Max:      c.Duration("max"),
		Verbose:  c.Bool("verbose"),
		DryRun:   c.Bool("dryrun"),
		Dump:     c.Bool("dump"),
	}

	if err := plugin.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
