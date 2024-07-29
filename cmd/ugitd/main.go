package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/utils/trace"
	"go.jolheiser.com/tailroute"
	"go.jolheiser.com/ugit/internal/git"
	"go.jolheiser.com/ugit/internal/http"
	"go.jolheiser.com/ugit/internal/ssh"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "pre-receive-hook" {
		preReceive()
		return
	}

	args, err := parseArgs(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		panic(err)
	}
	args.RepoDir, err = filepath.Abs(args.RepoDir)
	if err != nil {
		panic(err)
	}

	log.SetLevel(args.Log.Level)
	middleware.DefaultLogger = httplog.RequestLogger(httplog.NewLogger("ugit", httplog.Options{
		JSON:     args.Log.JSON,
		LogLevel: slog.Level(args.Log.Level),
		Concise:  args.Log.Level != log.DebugLevel,
	}))

	if args.Log.Level == log.DebugLevel {
		trace.SetTarget(trace.Packet)
	} else {
		middleware.DefaultLogger = http.NoopLogger
		ssh.DefaultLogger = ssh.NoopLogger
	}

	if args.Log.JSON {
		log.SetFormatter(log.JSONFormatter)
	}

	if err := requiredFS(args.RepoDir); err != nil {
		panic(err)
	}

	if args.SSH.Enable {
		sshSettings := ssh.Settings{
			AuthorizedKeys: args.SSH.AuthorizedKeys,
			CloneURL:       args.SSH.CloneURL,
			Port:           args.SSH.Port,
			HostKey:        args.SSH.HostKey,
			RepoDir:        args.RepoDir,
		}
		sshSrv, err := ssh.New(sshSettings)
		if err != nil {
			panic(err)
		}
		go func() {
			log.Debugf("SSH listening on ssh://localhost:%d\n", sshSettings.Port)
			if err := sshSrv.ListenAndServe(); err != nil {
				panic(err)
			}
		}()
	}

	httpSettings := http.Settings{
		Title:       args.Meta.Title,
		Description: args.Meta.Description,
		CloneURL:    args.HTTP.CloneURL,
		Port:        args.HTTP.Port,
		RepoDir:     args.RepoDir,
		Profile: http.Profile{
			Username: args.Profile.Username,
			Email:    args.Profile.Email,
		},
		ShowPrivate: false,
	}
	for _, link := range args.Profile.Links {
		httpSettings.Profile.Links = append(httpSettings.Profile.Links, http.Link{
			Name: link.Name,
			URL:  link.URL,
		})
	}
	if args.HTTP.Enable {
		httpSrv := http.New(httpSettings)
		go func() {
			log.Debugf("HTTP listening on http://localhost:%d\n", httpSettings.Port)
			if err := httpSrv.ListenAndServe(); err != nil {
				panic(err)
			}
		}()
	}

	if args.Tailscale.Enable {
		tailnetSettings := httpSettings
		tailnetSettings.ShowPrivate = true
		tailnetSrv := http.New(tailnetSettings)
		tr := tailroute.Router{
			Tailnet: tailnetSrv.Mux,
		}
		go func() {
			log.Debugf("Tailnet listening on http://%s\n", args.Tailscale.Hostname)
			if err := tr.Serve(args.Tailscale.Hostname, args.Tailscale.DataDir); err != nil {
				panic(err)
			}
		}()
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Kill, os.Interrupt)
	<-ch
}

func requiredFS(repoDir string) error {
	if err := os.MkdirAll(repoDir, os.ModePerm); err != nil {
		return err
	}

	if !git.RequiresHook {
		return nil
	}
	bin, err := os.Executable()
	if err != nil {
		return err
	}

	fp := filepath.Join(repoDir, "hooks")
	if err := os.MkdirAll(fp, os.ModePerm); err != nil {
		return err
	}
	fp = filepath.Join(fp, "pre-receive")

	if err := os.MkdirAll(fp+".d", os.ModePerm); err != nil {
		return err
	}

	fi, err := os.Create(fp)
	if err != nil {
		return err
	}
	fi.WriteString("#!/usr/bin/env bash\n")
	fi.WriteString(fmt.Sprintf("%s pre-receive-hook\n", bin))
	fi.WriteString(fmt.Sprintf(`for hook in %s.d/*; do
	"${hook}"
done`, fp))
	fi.Close()

	return os.Chmod(fp, 0o755)
}

func preReceive() {
	repoDir, ok := os.LookupEnv("UGIT_REPODIR")
	if !ok {
		panic("UGIT_REPODIR is not set")
	}

	opts := make([]*packp.Option, 0)
	if pushCount, err := strconv.Atoi(os.Getenv("GIT_PUSH_OPTION_COUNT")); err == nil {
		for idx := 0; idx < pushCount; idx++ {
			opt := os.Getenv(fmt.Sprintf("GIT_PUSH_OPTION_%d", idx))
			kv := strings.SplitN(opt, "=", 2)
			if len(kv) == 2 {
				opts = append(opts, &packp.Option{
					Key:   kv[0],
					Value: kv[1],
				})
			}
		}
	}

	repo, err := git.NewRepo(filepath.Dir(repoDir), filepath.Base(repoDir))
	if err != nil {
		panic(err)
	}
	if err := git.HandlePushOptions(repo, opts); err != nil {
		panic(err)
	}
}
