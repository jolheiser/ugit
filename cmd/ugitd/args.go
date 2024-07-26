package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffyaml"
)

type cliArgs struct {
	RepoDir   string
	SSH       sshArgs
	HTTP      httpArgs
	Meta      metaArgs
	Profile   profileArgs
	Log       logArgs
	Tailscale tailscaleArgs
}

type sshArgs struct {
	AuthorizedKeys string
	CloneURL       string
	Port           int
	HostKey        string
}

type httpArgs struct {
	CloneURL string
	Port     int
}

type metaArgs struct {
	Title       string
	Description string
}

type profileArgs struct {
	Username string
	Email    string
	Links    []profileLink
}

type profileLink struct {
	Name string
	URL  string
}

type logArgs struct {
	Level log.Level
	JSON  bool
}

type tailscaleArgs struct {
	Hostname string
	DataDir  string
}

func parseArgs(args []string) (c cliArgs, e error) {
	fs := flag.NewFlagSet("ugitd", flag.ContinueOnError)
	fs.String("config", "ugit.yaml", "Path to config file")

	c = cliArgs{
		RepoDir: ".ugit",
		SSH: sshArgs{
			AuthorizedKeys: ".ssh/authorized_keys",
			CloneURL:       "ssh://localhost:8448",
			Port:           8448,
			HostKey:        ".ssh/ugit_ed25519",
		},
		HTTP: httpArgs{
			CloneURL: "http://localhost:8449",
			Port:     8449,
		},
		Meta: metaArgs{
			Title:       "ugit",
			Description: "Minimal git server",
		},
		Log: logArgs{
			Level: log.InfoLevel,
		},
		Tailscale: tailscaleArgs{
			Hostname: "ugit",
			DataDir:  ".tsnet",
		},
	}

	fs.Func("log.level", "Logging level", func(s string) error {
		lvl, err := log.ParseLevel(s)
		if err != nil {
			return err
		}
		c.Log.Level = lvl
		return nil
	})
	fs.BoolVar(&c.Log.JSON, "log.json", c.Log.JSON, "Print logs in JSON(L) format")
	fs.StringVar(&c.RepoDir, "repo-dir", c.RepoDir, "Path to directory containing repositories")
	fs.StringVar(&c.SSH.AuthorizedKeys, "ssh.authorized-keys", c.SSH.AuthorizedKeys, "Path to authorized_keys")
	fs.StringVar(&c.SSH.CloneURL, "ssh.clone-url", c.SSH.CloneURL, "SSH clone URL base")
	fs.IntVar(&c.SSH.Port, "ssh.port", c.SSH.Port, "SSH port")
	fs.StringVar(&c.SSH.HostKey, "ssh.host-key", c.SSH.HostKey, "SSH host key (created if it doesn't exist)")
	fs.StringVar(&c.HTTP.CloneURL, "http.clone-url", c.HTTP.CloneURL, "HTTP clone URL base")
	fs.IntVar(&c.HTTP.Port, "http.port", c.HTTP.Port, "HTTP port")
	fs.StringVar(&c.Meta.Title, "meta.title", c.Meta.Title, "App title")
	fs.StringVar(&c.Meta.Description, "meta.description", c.Meta.Description, "App description")
	fs.StringVar(&c.Profile.Username, "profile.username", c.Profile.Username, "Username for index page")
	fs.StringVar(&c.Profile.Email, "profile.email", c.Profile.Email, "Email for index page")
	fs.Func("profile.links", "Link(s) for index page", func(s string) error {
		parts := strings.SplitN(s, ",", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid profile link %q", s)
		}
		c.Profile.Links = append(c.Profile.Links, profileLink{
			Name: parts[0],
			URL:  parts[1],
		})
		return nil
	})
	fs.StringVar(&c.Tailscale.Hostname, "tailscale.hostname", c.Tailscale.Hostname, "Tailscale host to show private repos on")
	fs.StringVar(&c.Tailscale.DataDir, "tailscale.data-dir", c.Tailscale.DataDir, "Tailscale data/state directory")

	return c, ff.Parse(fs, args,
		ff.WithEnvVarPrefix("UGIT"),
		ff.WithConfigFileFlag("config"),
		ff.WithAllowMissingConfigFile(true),
		ff.WithConfigFileParser(ffyaml.Parser),
	)
}
