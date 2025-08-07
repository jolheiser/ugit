package git

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// RepoMeta is the meta information a Repo can have
type RepoMeta struct {
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Tags        TagSet `json:"tags"`
}

// TagSet is a Set of tags
type TagSet map[string]struct{}

// Add adds a tag to the set
func (t TagSet) Add(tag string) {
	t[tag] = struct{}{}
}

// Remove removes a tag from the set
func (t TagSet) Remove(tag string) {
	delete(t, tag)
}

// Contains checks if a tag is in the set
func (t TagSet) Contains(tag string) bool {
	_, ok := t[tag]
	return ok
}

// Slice returns the set as a (sorted) slice
func (t TagSet) Slice() []string {
	s := make([]string, 0, len(t))
	for k := range t {
		s = append(s, k)
	}
	slices.Sort(s)
	return s
}

// MarshalJSON implements [json.Marshaler]
func (t TagSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Slice())
}

// UnmarshalJSON implements [json.Unmarshaler]
func (t *TagSet) UnmarshalJSON(b []byte) error {
	if *t == nil {
		ts := make(TagSet)
		t = &ts
	}
	var s []string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	for _, ss := range s {
		t.Add(ss)
	}
	return nil
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
