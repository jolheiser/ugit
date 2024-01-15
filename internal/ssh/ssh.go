package ssh

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
)

// Settings holds the configuration for the SSH server
type Settings struct {
	AuthorizedKeys string
	CloneURL       string
	Port           int
	HostKey        string
	RepoDir        string
}

// New creates a new SSH server.
func New(settings Settings) (*ssh.Server, error) {
	s, err := wish.NewServer(
		wish.WithAuthorizedKeys(settings.AuthorizedKeys),
		wish.WithAddress(fmt.Sprintf(":%d", settings.Port)),
		wish.WithHostKeyPath(settings.HostKey),
		wish.WithMiddleware(
			Middleware(settings.RepoDir, settings.CloneURL, settings.Port, hooks{}),
			logging.MiddlewareWithLogger(DefaultLogger),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create new SSH server: %w", err)
	}

	return s, nil
}

type hooks struct{}

func (a hooks) Push(_ string, _ ssh.PublicKey)  {}
func (a hooks) Fetch(_ string, _ ssh.PublicKey) {}

var (
	DefaultLogger logging.Logger = log.StandardLog()
	NoopLogger    logging.Logger = noopLogger{}
)

type noopLogger struct{}

func (n noopLogger) Printf(format string, v ...interface{}) {}
