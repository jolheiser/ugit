package html

import (
	"bytes"
	"path/filepath"

	"go.jolheiser.com/ugit/internal/git"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

var Markdown = goldmark.New(
	goldmark.WithRendererOptions(
		goldmarkhtml.WithUnsafe(),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithExtensions(
		extension.GFM,
		emoji.Emoji,
		highlighting.NewHighlighting(
			highlighting.WithStyle("catppuccin-mocha"),
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
				chromahtml.WithLineNumbers(true),
				chromahtml.WithLinkableLineNumbers(true, "md-"),
				chromahtml.LineNumbersInTable(true),
			),
		),
	),
)

func Readme(repo *git.Repo, ref, path string) (string, error) {
	var readme string
	var err error
	for _, md := range []string{"README.md", "readme.md"} {
		readme, err = repo.FileContent(ref, filepath.Join(path, md))
		if err == nil {
			break
		}
	}

	if readme != "" {
		var buf bytes.Buffer
		if err := Markdown.Convert([]byte(readme), &buf); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	for _, md := range []string{"README.txt", "README", "readme.txt", "readme"} {
		readme, err = repo.FileContent(ref, filepath.Join(path, md))
		if err == nil {
			return readme, nil
		}
	}

	return "", nil
}
