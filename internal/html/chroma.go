package html

import (
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

var (
	Formatter = html.New(
		html.WithLineNumbers(true),
		html.WithLinkableLineNumbers(true, "L"),
		html.WithClasses(true),
		html.LineNumbersInTable(true),
	)
	basicFormatter = html.New(
		html.WithClasses(true),
	)
	Code = code{}
)

type code struct{}

func (c code) setup(source []byte, fileName string) (chroma.Iterator, *chroma.Style, error) {
	lexer := lexers.Match(fileName)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("catppuccin-mocha")
	if style == nil {
		style = styles.Fallback
	}

	iter, err := lexer.Tokenise(nil, string(source))
	if err != nil {
		return nil, nil, err
	}

	return iter, style, nil
}

func (c code) Basic(source []byte, fileName string, writer io.Writer) error {
	iter, style, err := c.setup(source, fileName)
	if err != nil {
		return err
	}
	return basicFormatter.Format(writer, style, iter)
}

func (c code) Convert(source []byte, fileName string, writer io.Writer) error {
	iter, style, err := c.setup(source, fileName)
	if err != nil {
		return err
	}
	return Formatter.Format(writer, style, iter)
}
