package html

import (
	_ "embed"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type RepoFileContext struct {
	BaseContext
	RepoHeaderComponentContext
	RepoBreadcrumbComponentContext
	Code string
}

//go:embed repo_file.js
var repoFileJS string

func RepoFileTemplate(rfc RepoFileContext) Node {
	return base(rfc.BaseContext, []Node{
		repoHeaderComponent(rfc.RepoHeaderComponentContext),
		Div(Class("mt-2 text-text"),
			repoBreadcrumbComponent(rfc.RepoBreadcrumbComponentContext),
			Text(" - "),
			A(Class("text-text underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href("?raw"), Text("raw")),
			Div(Class("code relative"),
				Raw(rfc.Code),
				Button(ID("copy"), Class("absolute top-0 right-0 rounded bg-base hover:bg-surface0")),
			),
		),
		Script(Raw(repoFileJS)),
	}...)
}
