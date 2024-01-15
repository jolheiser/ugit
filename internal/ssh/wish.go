package ssh

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.jolheiser.com/ugit/internal/git"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

// ErrSystemMalfunction represents a general system error returned to clients.
var ErrSystemMalfunction = errors.New("something went wrong")

// ErrInvalidRepo represents an attempt to access a non-existent repo.
var ErrInvalidRepo = errors.New("invalid repo")

// Hooks is an interface that allows for custom authorization
// implementations and post push/fetch notifications. Prior to git access,
// AuthRepo will be called with the ssh.Session public key and the repo name.
// Implementers return the appropriate AccessLevel.
type Hooks interface {
	Push(string, ssh.PublicKey)
	Fetch(string, ssh.PublicKey)
}

// Session wraps sn ssh.Session to implement git.ReadWriteContexter
type Session struct {
	s ssh.Session
}

// Read implements io.Reader
func (s Session) Read(p []byte) (n int, err error) {
	return s.s.Read(p)
}

// Write implements io.Writer
func (s Session) Write(p []byte) (n int, err error) {
	return s.s.Write(p)
}

// Close implements io.Closer
func (s Session) Close() error {
	return nil
}

// Context returns an interface context.Context
func (s Session) Context() context.Context {
	return s.s.Context()
}

// Middleware adds Git server functionality to the ssh.Server. Repos are stored
// in the specified repo directory. The provided Hooks implementation will be
// checked for access on a per repo basis for a ssh.Session public key.
// Hooks.Push and Hooks.Fetch will be called on successful completion of
// their commands.
func Middleware(repoDir string, cloneURL string, port int, gh Hooks) wish.Middleware {
	return func(sh ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			sess := Session{s: s}
			cmd := s.Command()

			// Git operations
			if len(cmd) == 2 {
				gc := cmd[0]
				// repo should be in the form of "repo.git" or "user/repo.git"
				repo := strings.TrimSuffix(strings.TrimPrefix(cmd[1], "/"), "/")
				repo = filepath.Clean(repo)
				if n := strings.Count(repo, "/"); n > 1 {
					Fatal(s, ErrInvalidRepo)
					return
				}
				pk := s.PublicKey()
				switch gc {
				case "git-receive-pack":
					if err := gitPack(sess, gc, repoDir, repo); err != nil {
						Fatal(s, ErrSystemMalfunction)
					}
					gh.Push(repo, pk)
					return
				case "git-upload-archive", "git-upload-pack":
					if err := gitPack(sess, gc, repoDir, repo); err != nil {
						if errors.Is(err, ErrInvalidRepo) {
							Fatal(s, ErrInvalidRepo)
						}
						log.Error("unknown git error", "error", err)
						Fatal(s, ErrSystemMalfunction)
					}
					gh.Fetch(repo, pk)
					return
				}
			}

			// Repo list
			if len(cmd) == 0 {
				des, err := os.ReadDir(repoDir)
				if err != nil && err != fs.ErrNotExist {
					log.Error("invalid repository", "error", err)
				}
				for _, de := range des {
					fmt.Fprintln(s, de.Name())
					fmt.Fprintf(s, "\tgit clone %s/%s\n", cloneURL, de.Name())
				}
			}
			sh(s)
		}
	}
}

func gitPack(s Session, gitCmd string, repoDir string, repoName string) error {
	rp := filepath.Join(repoDir, repoName)
	protocol, err := git.NewProtocol(rp)
	if err != nil {
		return err
	}
	switch gitCmd {
	case "git-upload-pack":
		exists, err := git.PathExists(rp)
		if !exists {
			return ErrInvalidRepo
		}
		if err != nil {
			return err
		}
		return protocol.SSHUploadPack(s)
	case "git-receive-pack":
		err := git.EnsureRepo(repoDir, repoName)
		if err != nil {
			return err
		}
		repo, err := git.NewRepo(repoDir, repoName)
		if err != nil {
			return err
		}
		err = protocol.SSHReceivePack(s, repo)
		if err != nil {
			return err
		}
		_, err = repo.DefaultBranch()
		if err != nil {
			return err
		}
		// Needed for git dumb http server
		return git.UpdateServerInfo(rp)
	default:
		return fmt.Errorf("unknown git command: %s", gitCmd)
	}
}

// Fatal prints to the session's STDOUT as a git response and exit 1.
func Fatal(s ssh.Session, v ...interface{}) {
	msg := fmt.Sprint(v...)
	// hex length includes 4 byte length prefix and ending newline
	pktLine := fmt.Sprintf("%04x%s\n", len(msg)+5, msg)
	_, _ = wish.WriteString(s, pktLine)
	s.Exit(1) // nolint: errcheck
}
