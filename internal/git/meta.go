package git

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// RepoMeta is the meta information a Repo can have
type RepoMeta struct {
	Description string   `json:"description"`
	Private     bool     `json:"private"`
	Tags        []string `json:"tags"`
}

// Update updates meta given another RepoMeta
func (m *RepoMeta) Update(meta RepoMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, m)
}

func (r Repo) metaPath() string {
	return filepath.Join(r.path, "ugit.json")
}

// SaveMeta saves the meta info of a Repo
func (r Repo) SaveMeta() error {
	// Compatibility with gitweb, because why not
	// Ignoring the error because it's not technically detrimental to ugit
	desc, err := os.Create(filepath.Join(r.path, "description"))
	if err == nil {
		defer desc.Close()
		_, _ = desc.WriteString(r.Meta.Description)
	}

	fi, err := os.Create(r.metaPath())
	if err != nil {
		return err
	}
	defer fi.Close()
	return json.NewEncoder(fi).Encode(r.Meta)
}

var defaultMeta = func() []byte {
	b, err := json.Marshal(RepoMeta{
		Private: true,
	})
	if err != nil {
		panic(fmt.Sprintf("could not init default meta: %v", err))
	}
	return b
}()

func ensureJSONFile(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	fi, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fi.Close()
	if _, err := fi.Write(defaultMeta); err != nil {
		return err
	}
	return nil
}
