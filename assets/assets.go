package assets

import "embed"

var (
	//go:embed *.svg
	Icons     embed.FS
	LinkIcon  = must("link.svg")
	EmailIcon = must("email.svg")
	LogoIcon  = must("ugit.svg")
)

func must(path string) []byte {
	content, err := Icons.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return content
}
