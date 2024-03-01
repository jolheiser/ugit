package markup

import (
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

var (
	// Formatter is the default formatter
	Formatter = html.New(
		html.WithLineNumbers(true),
		html.WithLinkableLineNumbers(true, "L"),
		html.WithClasses(true),
		html.LineNumbersInTable(true),
	)
	basicFormatter = html.New(
		html.WithClasses(true),
	)
	// Code is the entrypoint for formatting
	Code = code{}
)

type code struct{}

func setup(source []byte, fileName string) (chroma.Iterator, *chroma.Style, error) {
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

// Basic formats code without any extras
func (c code) Basic(source []byte, fileName string, writer io.Writer) error {
	iter, style, err := setup(source, fileName)
	if err != nil {
		return err
	}
	return basicFormatter.Format(writer, style, iter)
}

// Convert formats code with line numbers, links, etc.
func (c code) Convert(source []byte, fileName string, writer io.Writer) error {
	iter, style, err := setup(source, fileName)
	if err != nil {
		return err
	}
	return Formatter.Format(writer, style, iter)
}

// Snippet formats code with line numbers starting at a specific line
func Snippet(source []byte, fileName string, line int, writer io.Writer) error {
	iter, style, err := setup(source, fileName)
	if err != nil {
		return err
	}
	formatter := html.New(
		html.WithLineNumbers(true),
		html.WithClasses(true),
		html.LineNumbersInTable(true),
		html.BaseLineNumber(line),
	)
	return formatter.Format(writer, style, iter)
}
