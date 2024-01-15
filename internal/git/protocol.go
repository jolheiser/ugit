package git

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/serverinfo"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/utils/ioutil"
)

type ReadWriteContexter interface {
	io.ReadWriteCloser
	Context() context.Context
}

type Protocol struct {
	endpoint *transport.Endpoint
	server   transport.Transport
}

func NewProtocol(repoPath string) (Protocol, error) {
	endpoint, err := transport.NewEndpoint("/")
	if err != nil {
		return Protocol{}, err
	}
	fs := osfs.New(repoPath)
	loader := server.NewFilesystemLoader(fs)
	gitServer := server.NewServer(loader)
	return Protocol{
		endpoint: endpoint,
		server:   gitServer,
	}, nil
}

func (p Protocol) HTTPInfoRefs(rwc ReadWriteContexter) error {
	session, err := p.server.NewUploadPackSession(p.endpoint, nil)
	if err != nil {
		return err
	}
	defer ioutil.CheckClose(rwc, &err)
	return p.infoRefs(rwc, session, "# service=git-upload-pack")
}

func (p Protocol) infoRefs(rwc ReadWriteContexter, session transport.UploadPackSession, prefix string) error {
	ar, err := session.AdvertisedReferencesContext(rwc.Context())
	if err != nil {
		return err
	}

	if prefix != "" {
		ar.Prefix = [][]byte{
			[]byte(prefix),
			pktline.Flush,
		}
	}

	if err := ar.Encode(rwc); err != nil {
		return err
	}

	return nil
}

func (p Protocol) HTTPUploadPack(rwc ReadWriteContexter) error {
	return p.uploadPack(rwc, false)
}

func (p Protocol) SSHUploadPack(rwc ReadWriteContexter) error {
	return p.uploadPack(rwc, true)
}

func (p Protocol) uploadPack(rwc ReadWriteContexter, ssh bool) error {
	session, err := p.server.NewUploadPackSession(p.endpoint, nil)
	if err != nil {
		return err
	}
	defer ioutil.CheckClose(rwc, &err)

	if ssh {
		if err := p.infoRefs(rwc, session, ""); err != nil {
			return err
		}
	}

	req := packp.NewUploadPackRequest()
	if err := req.Decode(rwc); err != nil {
		return err
	}

	var resp *packp.UploadPackResponse
	resp, err = session.UploadPack(rwc.Context(), req)
	if err != nil {
		return err
	}

	if err := resp.Encode(rwc); err != nil {
		return fmt.Errorf("could not encode upload pack: %w", err)
	}

	return nil
}

func (p Protocol) SSHReceivePack(rwc ReadWriteContexter, repo *Repo) error {
	buf := bufio.NewReader(rwc)

	session, err := p.server.NewReceivePackSession(p.endpoint, nil)
	if err != nil {
		return err
	}

	ar, err := session.AdvertisedReferencesContext(rwc.Context())
	if err != nil {
		return fmt.Errorf("internal error in advertised references: %w", err)
	}
	_ = ar.Capabilities.Set(capability.PushOptions)
	_ = ar.Capabilities.Set("no-thin")

	if err := ar.Encode(rwc); err != nil {
		return fmt.Errorf("error in advertised references encoding: %w", err)
	}

	req := packp.NewReferenceUpdateRequest()
	_ = req.Capabilities.Set(capability.ReportStatus)
	if err := req.Decode(buf); err != nil {
		// FIXME this is a hack, but go-git doesn't accept a 0000 if there are no refs to update
		if !strings.EqualFold(err.Error(), "capabilities delimiter not found") {
			return fmt.Errorf("error decoding: %w", err)
		}
	}

	// FIXME also a hack, if the next bytes are PACK then we have a packfile, otherwise assume it's push options
	peek, err := buf.Peek(4)
	if err != nil {
		return err
	}
	if string(peek) != "PACK" {
		s := pktline.NewScanner(buf)
		for s.Scan() {
			val := string(s.Bytes())
			if val == "" {
				break
			}
			if s.Err() != nil {
				return s.Err()
			}
			parts := strings.SplitN(val, "=", 2)
			req.Options = append(req.Options, &packp.Option{
				Key:   parts[0],
				Value: parts[1],
			})
		}
	}

	if err := handlePushOptions(repo, req.Options); err != nil {
		return fmt.Errorf("could not handle push options: %w", err)
	}

	// FIXME if there are only delete commands, there is no packfile and ReceivePack will block forever
	noPack := true
	for _, c := range req.Commands {
		if c.Action() != packp.Delete {
			noPack = false
			break
		}
	}
	if noPack {
		req.Packfile = nil
	}

	rs, err := session.ReceivePack(rwc.Context(), req)
	if err != nil {
		return fmt.Errorf("error in receive pack: %w", err)
	}

	if err := rs.Encode(rwc); err != nil {
		return fmt.Errorf("could not encode receive pack: %w", err)
	}

	return nil
}

func handlePushOptions(repo *Repo, opts []*packp.Option) error {
	var changed bool
	for _, opt := range opts {
		switch strings.ToLower(opt.Key) {
		case "desc", "description":
			changed = repo.Meta.Description != opt.Value
			repo.Meta.Description = opt.Value
		case "private":
			private, err := strconv.ParseBool(opt.Value)
			if err != nil {
				continue
			}
			changed = repo.Meta.Private != private
			repo.Meta.Private = private
		}
	}
	if changed {
		return repo.SaveMeta()
	}
	return nil
}

func UpdateServerInfo(repo string) error {
	r, err := git.PlainOpen(repo)
	if err != nil {
		return err
	}
	fs := r.Storer.(*filesystem.Storage).Filesystem()
	return serverinfo.UpdateServerInfo(r.Storer, fs)
}
