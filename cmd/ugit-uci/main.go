package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-git/go-git/v5"
)

func maine() error {
	repoDir, ok := os.LookupEnv("UGIT_REPODIR")
	if !ok {
		panic("UGIT_REPODIR not set")
	}

	urlFlag := flag.String("url", "http://localhost:3448", "URL for UCI")
	flag.StringVar(urlFlag, "u", *urlFlag, "--url")
	repoDirFlag := flag.String("repo-dir", "", "Repo dir (including .git)")
	flag.StringVar(repoDirFlag, "rd", *repoDirFlag, "--repo-dir")
	manifestFlag := flag.String("manifest", ".uci.jsonnet", "Path to manifest in repo")
	flag.StringVar(manifestFlag, "m", *manifestFlag, "--manifest")
	flag.Parse()

	if *repoDirFlag != "" {
		repoDir = *repoDirFlag
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return fmt.Errorf("could not open git dir %q: %w", repoDir, err)
	}

	tree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("could not get worktree: %w", err)
	}

	fi, err := tree.Filesystem.Open(*manifestFlag)
	if err != nil {
		return fmt.Errorf("could not open manifest %q: %w", *manifestFlag, err)
	}
	defer fi.Close()

	data, err := io.ReadAll(fi)
	if err != nil {
		return fmt.Errorf("could not read manifest: %w", err)
	}

	resp, err := http.Post(*urlFlag, "application/jsonnet", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("could not post manifest: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-ok response: %s", resp.Status)
	}

	return nil
}

func main() {
	if err := maine(); err != nil {
		fmt.Println(err)
	}
}
