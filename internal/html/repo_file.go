package html

import (
	_ "embed"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type RepoFileContext struct {
	BaseContext
	RepoHeaderComponentContext
	RepoBreadcrumbComponentContext
	Code   string
	Commit string
	Path   string
}

//go:embed repo_file.js
var repoFileJS string

func RepoFileTemplate(rfc RepoFileContext) Node {
	permalink := fmt.Sprintf("/%s/tree/%s/%s", rfc.RepoBreadcrumbComponentContext.Repo, rfc.Commit, rfc.Path)
	return base(rfc.BaseContext, []Node{
		repoHeaderComponent(rfc.RepoHeaderComponentContext),
		Div(Class("mt-2 text-text"),
			repoBreadcrumbComponent(rfc.RepoBreadcrumbComponentContext),
			Text(" - "),
			A(Class("text-text underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href("?raw"), Text("raw")),
			Text(" - "),
			A(Class("text-text underline decoration-text/50 decoration-dashed hover:decoration-solid"), ID("permalink"), Data("permalink", permalink), Href(permalink), Text("permalink")),
			Div(Class("code relative"),
				Raw(rfc.Code),
				Button(ID("copy"), Class("absolute top-0 right-0 rounded bg-base hover:bg-surface0")),
			),
		),
		Script(Raw(repoFileJS)),
	}...)
}
