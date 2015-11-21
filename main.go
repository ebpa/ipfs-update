package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	cli "github.com/codegangsta/cli"
	stump "github.com/whyrusleeping/stump"
)

var (
	globalGatewayUrl = "https://ipfs.io"
	localApiUrl      = "http://localhost:5001"
	ipfsVersionPath  = "/ipfs/QmSiTko9JZyabH56y2fussEt1A5oDqsFXB3CkvAqraFryz"
)

func main() {
	app := cli.NewApp()
	app.Author = "whyrusleeping"
	app.Usage = "update ipfs"
	app.Version = "0.1.0"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "print verbose output",
		},
	}

	app.Before = func(c *cli.Context) error {
		stump.Verbose = c.Bool("verbose")
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  "versions",
			Usage: "print out all available versions",
			Action: func(c *cli.Context) {
				vs, err := GetVersions(ipfsVersionPath)
				if err != nil {
					stump.Fatal("Failed to query versions: ", err)
				}

				for _, v := range vs {
					fmt.Println(v)
				}
			},
		},
		{
			Name:  "version",
			Usage: "print out currently installed version",
			Action: func(c *cli.Context) {
				v, err := GetCurrentVersion()
				if err != nil {
					stump.Fatal("Failed to check local version: ", err)
				}

				fmt.Println(v)
			},
		},
		{
			Name:  "install",
			Usage: "install a version of ipfs",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "no-check",
					Usage: "skip running of pre-install tests",
				},
			},
			Action: func(c *cli.Context) {
				vers := c.Args().First()
				if vers == "" {
					stump.Fatal("Please specify a version to install")
				}
				if vers == "latest" {
					latest, err := GetLatestVersion(ipfsVersionPath)
					if err != nil {
						stump.Fatal("error resolving 'latest': ", err)
					}
					vers = latest
				}

				err := InstallVersion(ipfsVersionPath, vers, c.Bool("no-check"))
				if err != nil {
					stump.Fatal(err)
				}
				stump.Log("\ninstallation complete.")

				if hasDaemonRunning() {
					stump.Log("remember to restart your daemon before continuing")
				}
			},
		},
		{
			Name:  "stash",
			Usage: "stashes copy of currently installed ipfs binary",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tag",
					Usage: "optionally specify tag for stashed binary",
				},
			},
			Action: func(c *cli.Context) {
				tag := c.String("tag")
				if tag == "" {
					vers, err := GetCurrentVersion()
					if err != nil {
						stump.Fatal(err)
					}
					tag = vers
				}

				_, err := StashOldBinary(tag, true)
				if err != nil {
					stump.Fatal(err)
				}
			},
		},
		{
			Name:  "revert",
			Usage: "revert to previously installed version of ipfs",
			Description: `revert will check if a previous update left a stashed
binary and overwrite the current ipfs binary with it.`,
			Action: func(c *cli.Context) {
				oldbinpath, err := selectRevertBin()
				if err != nil {
					stump.Fatal(err)
				}

				oldpath, err := ioutil.ReadFile(filepath.Join(ipfsDir(), "old-bin", "path-old"))
				if err != nil {
					stump.Fatal("Path for previous installation could not be read: ", err)
				}

				binpath := string(oldpath)
				err = InstallBinaryTo(oldbinpath, binpath)
				if err != nil {
					stump.Error("failed to move old binary: %s", oldbinpath)
					stump.Error("to path: %s", binpath)
					stump.Fatal(err)
				}
			},
		},
		{
			Name:  "fetch",
			Usage: "fetch a given (default: latest) version of ipfs",
			Action: func(c *cli.Context) {
				vers := c.Args().First()
				if vers == "" || vers == "latest" {
					latest, err := GetLatestVersion(ipfsVersionPath)
					if err != nil {
						stump.Fatal("error querying latest version: ", err)
					}

					vers = latest
				}

				output := "ipfs-" + vers
				ofl := c.String("output")
				if ofl != "" {
					output = ofl
				}

				_, err := os.Stat(output)
				if err == nil {
					stump.Fatal("file named %s already exists")
				}

				if !os.IsNotExist(err) {
					stump.Fatal("stat(%s)", output, err)
				}

				err = GetBinaryForVersion(ipfsVersionPath, vers, output)
				if err != nil {
					stump.Fatal("Failed to fetch binary: ", err)
				}

				err = os.Chmod(output, 0755)
				if err != nil {
					stump.Fatal("setting new binary executable: ", err)
				}
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output",
					Usage: "specify where to save the downloaded binary",
				},
			},
		},
	}

	app.Run(os.Args)
}
