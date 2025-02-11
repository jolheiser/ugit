package main

import (
	"flag"
	"fmt"
	"log/slog"
	"strings"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffyaml"
)

type cliArgs struct {
	RepoDir     string
	SSH         sshArgs
	HTTP        httpArgs
	Meta        metaArgs
	Profile     profileArgs
	Log         logArgs
	ShowPrivate bool
}

type sshArgs struct {
	Enable         bool
	AuthorizedKeys string
	CloneURL       string
	Port           int
	HostKey        string
}

type httpArgs struct {
	Enable   bool
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
	Level slog.Level
	JSON  bool
}

func parseArgs(args []string) (c cliArgs, e error) {
	fs := flag.NewFlagSet("ugitd", flag.ContinueOnError)
	fs.String("config", "ugit.yaml", "Path to config file")

	c = cliArgs{
		RepoDir: ".ugit",
		SSH: sshArgs{
			Enable:         true,
			AuthorizedKeys: ".ssh/authorized_keys",
			CloneURL:       "ssh://localhost:8448",
			Port:           8448,
			HostKey:        ".ssh/ugit_ed25519",
		},
		HTTP: httpArgs{
			Enable:   true,
			CloneURL: "http://localhost:8449",
			Port:     8449,
		},
		Meta: metaArgs{
			Title:       "ugit",
			Description: "Minimal git server",
		},
		Log: logArgs{
			Level: slog.LevelError,
		},
	}

	fs.Func("log.level", "Logging level", func(s string) error {
		var lvl slog.Level
		switch strings.ToLower(s) {
		case "debug":
			lvl = slog.LevelDebug
		case "info":
			lvl = slog.LevelInfo
		case "warn", "warning":
			lvl = slog.LevelWarn
		case "error":
			lvl = slog.LevelError
		default:
			return fmt.Errorf("unknown log level %q: options are [debug, info, warn, error]", s)
		}
		c.Log.Level = lvl
		return nil
	})
	fs.BoolVar(&c.Log.JSON, "log.json", c.Log.JSON, "Print logs in JSON(L) format")
	fs.StringVar(&c.RepoDir, "repo-dir", c.RepoDir, "Path to directory containing repositories")
	fs.BoolVar(&c.SSH.Enable, "ssh.enable", c.SSH.Enable, "Enable SSH server")
	fs.StringVar(&c.SSH.AuthorizedKeys, "ssh.authorized-keys", c.SSH.AuthorizedKeys, "Path to authorized_keys")
	fs.StringVar(&c.SSH.CloneURL, "ssh.clone-url", c.SSH.CloneURL, "SSH clone URL base")
	fs.IntVar(&c.SSH.Port, "ssh.port", c.SSH.Port, "SSH port")
	fs.StringVar(&c.SSH.HostKey, "ssh.host-key", c.SSH.HostKey, "SSH host key (created if it doesn't exist)")
	fs.BoolVar(&c.HTTP.Enable, "http.enable", c.HTTP.Enable, "Enable HTTP server")
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

	return c, ff.Parse(fs, args,
		ff.WithEnvVarPrefix("UGIT"),
		ff.WithConfigFileFlag("config"),
		ff.WithAllowMissingConfigFile(true),
		ff.WithConfigFileParser(ffyaml.Parser),
	)
}
