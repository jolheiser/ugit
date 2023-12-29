package ssh

import (
	"fmt"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/git"
	"github.com/charmbracelet/wish/logging"
)

func New() (*ssh.Server, error) {
	s, err := wish.NewServer(
		wish.WithAuthorizedKeys(".ssh/authorized_keys"),
		wish.WithAddress("localhost:8448"),
		wish.WithHostKeyPath(".ssh/ugit_ed25519"),
		wish.WithMiddleware(
			git.Middleware(".ugit", app{}),
			logging.Middleware(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create new SSH server: %w", err)
	}

	return s, nil
}

type app struct{}

func (a app) AuthRepo(repo string, pk ssh.PublicKey) git.AccessLevel {
	return git.ReadWriteAccess
}
func (a app) Push(_ string, _ ssh.PublicKey)  {}
func (a app) Fetch(_ string, _ ssh.PublicKey) {}
