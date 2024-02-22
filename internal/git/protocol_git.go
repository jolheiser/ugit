//go:build !gogit

package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
)

var RequiresHook = true

type CmdProtocol string

func NewProtocol(repoPath string) (Protocoler, error) {
	return CmdProtocol(repoPath), nil
}

func (c CmdProtocol) HTTPInfoRefs(ctx ReadWriteContexter) error {
	pkt := pktline.NewEncoder(ctx)
	if err := pkt.EncodeString("# service=git-upload-pack"); err != nil {
		return err
	}
	if err := pkt.Flush(); err != nil {
		return err
	}
	return gitService(ctx, "upload-pack", string(c), "--stateless-rpc", "--advertise-refs")
}

func (c CmdProtocol) HTTPUploadPack(ctx ReadWriteContexter) error {
	return gitService(ctx, "upload-pack", string(c), "--stateless-rpc")
}

func (c CmdProtocol) SSHUploadPack(ctx ReadWriteContexter) error {
	return gitService(ctx, "upload-pack", string(c))
}

func (c CmdProtocol) SSHReceivePack(ctx ReadWriteContexter, _ *Repo) error {
	return gitService(ctx, "receive-pack", string(c))
}

func gitService(ctx ReadWriteContexter, command, repoDir string, args ...string) error {
	cmd := exec.CommandContext(ctx.Context(), "git")
	cmd.Args = append(cmd.Args, []string{
		"-c", "protocol.version=2",
		"-c", "uploadpack.allowFilter=true",
		"-c", "receive.advertisePushOptions=true",
		"-c", fmt.Sprintf("core.hooksPath=%s", filepath.Join(filepath.Dir(repoDir), "hooks")),
		command,
	}...)
	if len(args) > 0 {
		cmd.Args = append(cmd.Args, args...)
	}
	cmd.Args = append(cmd.Args, repoDir)
	cmd.Env = append(os.Environ(), fmt.Sprintf("UGIT_REPODIR=%s", repoDir), "GIT_PROTOCOL=version=2")
	cmd.Stdin = ctx
	cmd.Stdout = ctx
	fmt.Println(cmd.Env, cmd.String())

	return cmd.Run()
}
