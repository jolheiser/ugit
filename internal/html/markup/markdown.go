package markup

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"

	"go.jolheiser.com/ugit/internal/git"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var markdown = goldmark.New(
	goldmark.WithRendererOptions(
		goldmarkhtml.WithUnsafe(),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
		parser.WithASTTransformers(
			util.Prioritized(astTransformer{}, 100),
		),
	),
	goldmark.WithExtensions(
		extension.GFM,
		emoji.Emoji,
		highlighting.NewHighlighting(
			highlighting.WithStyle("catppuccin-mocha"),
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
			),
		),
	),
)

// Readme transforms a readme, potentially from markdown, into HTML
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
		ctx := parser.NewContext()
		mdCtx := markdownContext{
			repo: repo.Name(),
			ref:  ref,
			path: path,
		}
		ctx.Set(renderContextKey, mdCtx)
		var buf bytes.Buffer
		if err := markdown.Convert([]byte(readme), &buf, parser.WithContext(ctx)); err != nil {
			return "", err
		}
		var out bytes.Buffer
		if err := postProcess(buf.String(), mdCtx, &out); err != nil {
			return "", err
		}

		return out.String(), nil
	}

	for _, md := range []string{"README.txt", "README", "readme.txt", "readme"} {
		readme, err = repo.FileContent(ref, filepath.Join(path, md))
		if err == nil {
			return readme, nil
		}
	}

	return "", nil
}

var renderContextKey = parser.NewContextKey()

type markdownContext struct {
	repo string
	ref  string
	path string
}

type astTransformer struct{}

// Transform does two main things
// 1. Changes images to work relative to the source and wraps them in links
// 2. Changes links to work relative to the source
func (a astTransformer) Transform(node *ast.Document, _ text.Reader, pc parser.Context) {
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		ctx := pc.Get(renderContextKey).(markdownContext)

		switch v := n.(type) {
		case *ast.Image:
			link := v.Destination
			if len(link) > 0 && !bytes.HasPrefix(link, []byte("http")) {
				v.SetAttributeString("style", []byte("max-width:100%;"))
				v.Destination = []byte(resolveLink(ctx.repo, ctx.ref, ctx.path, string(link)) + "?raw&pretty")
			}

			parent := n.Parent()
			if _, ok := parent.(*ast.Link); !ok && parent != nil {
				next := n.NextSibling()
				wrapper := ast.NewLink()
				wrapper.Destination = v.Destination
				wrapper.Title = v.Title
				wrapper.SetAttributeString("target", []byte("_blank"))
				img := ast.NewImage(ast.NewLink())
				img.Destination = link
				img.Title = v.Title
				for _, attr := range v.Attributes() {
					img.SetAttribute(attr.Name, attr.Value)
				}
				for child := v.FirstChild(); child != nil; {
					nextChild := child.NextSibling()
					img.AppendChild(img, child)
					child = nextChild
				}
				wrapper.AppendChild(wrapper, img)
				wrapper.SetNextSibling(next)
				parent.ReplaceChild(parent, n, wrapper)
				v.SetNextSibling(next)
			}
		case *ast.Link:
			link := v.Destination
			if len(link) > 0 && !bytes.HasPrefix(link, []byte("http")) && link[0] != '#' && !bytes.HasPrefix(link, []byte("mailto")) {
				v.Destination = []byte(resolveLink(ctx.repo, ctx.ref, ctx.path, string(link)))
			}
		}

		return ast.WalkContinue, nil
	})
}

func postProcess(in string, ctx markdownContext, out io.Writer) error {
	node, err := html.Parse(strings.NewReader("<html><body>" + in + "</body></html"))
	if err != nil {
		return err
	}
	if node.Type == html.DocumentNode {
		node = node.FirstChild
	}

	process(ctx, node)

	renderNodes := make([]*html.Node, 0)
	if node.Data == "html" {
		node = node.FirstChild
		for node != nil && node.Data != "body" {
			node = node.NextSibling
		}
	}
	if node != nil {
		if node.Data == "body" {
			child := node.FirstChild
			for child != nil {
				renderNodes = append(renderNodes, child)
				child = child.NextSibling
			}
		} else {
			renderNodes = append(renderNodes, node)
		}
	}
	for _, node := range renderNodes {
		if err := html.Render(out, node); err != nil {
			return err
		}
	}
	return nil
}

func process(ctx markdownContext, node *html.Node) {
	if node.Type == html.ElementNode && node.Data == "img" {
		for i, attr := range node.Attr {
			if attr.Key != "src" {
				continue
			}
			if len(attr.Val) > 0 && !strings.HasPrefix(attr.Val, "http") && !strings.HasPrefix(attr.Val, "data:image/") {
				attr.Val = resolveLink(ctx.repo, ctx.ref, ctx.path, attr.Val) + "?raw&pretty"
			}
			node.Attr[i] = attr
		}
	}
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		process(ctx, n)
	}
}

func resolveLink(repo, ref, path, link string) string {
	baseURL, err := url.Parse(fmt.Sprintf("/%s/tree/%s/%s", repo, ref, path))
	if err != nil {
		return ""
	}
	linkURL, err := url.Parse(link)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(linkURL).String()
}
