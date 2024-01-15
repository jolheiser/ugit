package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"go.jolheiser.com/ugit/internal/http"
	"go.jolheiser.com/ugit/internal/ssh"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-git/go-git/v5/utils/trace"
)

func main() {
	args, err := parseArgs(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		panic(err)
	}

	if args.Debug {
		trace.SetTarget(trace.Packet)
		log.SetLevel(log.DebugLevel)
	} else {
		middleware.DefaultLogger = http.NoopLogger
		ssh.DefaultLogger = ssh.NoopLogger
	}

	if err := os.MkdirAll(args.RepoDir, os.ModePerm); err != nil {
		panic(err)
	}

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
		fmt.Printf("SSH listening on ssh://localhost:%d\n", sshSettings.Port)
		if err := sshSrv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

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
	}
	for _, link := range args.Profile.Links {
		httpSettings.Profile.Links = append(httpSettings.Profile.Links, http.Link{
			Name: link.Name,
			URL:  link.URL,
		})
	}
	httpSrv := http.New(httpSettings)
	go func() {
		fmt.Printf("HTTP listening on http://localhost:%d\n", httpSettings.Port)
		if err := httpSrv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Kill, os.Interrupt)
	<-ch
}
