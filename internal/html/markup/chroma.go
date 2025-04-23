package markup

import (
	"io"
	"path/filepath"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

var customReg = map[string]string{
	".hujson": "json",
}

// Options are the default set of formatting options
func Options(linePrefix string) []html.Option {
	return []html.Option{
		html.WithLineNumbers(true),
		html.WithLinkableLineNumbers(true, linePrefix),
		html.WithClasses(true),
		html.LineNumbersInTable(true),
	}
}

func setup(source []byte, fileName string) (chroma.Iterator, *chroma.Style, error) {
	lexer := lexers.Match(fileName)
	if lexer == nil {
		lexer = lexers.Fallback
		if name, ok := customReg[filepath.Ext(fileName)]; ok {
			lexer = lexers.Get(name)
		}
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

// Convert formats code with line numbers, links, etc.
func Convert(source []byte, fileName, linePrefix string, writer io.Writer) error {
	iter, style, err := setup(source, fileName)
	if err != nil {
		return err
	}
	return html.New(Options(linePrefix)...).Format(writer, style, iter)
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
