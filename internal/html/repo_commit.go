package html

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"go.jolheiser.com/ugit/internal/git"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type RepoCommitContext struct {
	BaseContext
	RepoHeaderComponentContext
	Commit git.Commit
}

func RepoCommitTemplate(rcc RepoCommitContext) Node {
	return base(rcc.BaseContext, []Node{
		repoHeaderComponent(rcc.RepoHeaderComponentContext),
		Div(Class("text-text mt-5"),
			A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/tree/%s/", rcc.RepoHeaderComponentContext.Name, rcc.Commit.SHA)), Text("tree")),
			Text(" "),
			A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/log/%s", rcc.RepoHeaderComponentContext.Name, rcc.Commit.SHA)), Text("log")),
			Text(" "),
			A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/commit/%s.patch", rcc.RepoHeaderComponentContext.Name, rcc.Commit.SHA)), Text("patch")),
		),
		Div(Class("text-text whitespace-pre mt-5 p-3 bg-base rounded"), Text(rcc.Commit.Message)),
		If(rcc.Commit.Signature != "",
			Details(Class("text-text whitespace-pre"),
				Summary(Class("cursor-pointer"), Text("Signature")),
				Div(Class("p-3 bg-base rounded"),
					Code(Text(rcc.Commit.Signature)),
				),
			),
		),
		Div(Class("text-text mt-3"),
			Div(
				Text(rcc.Commit.Author),
				Text(" "),
				A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("mailto:%s", rcc.Commit.Email)), Text(fmt.Sprintf("<%s>", rcc.Commit.Email))),
			),
			Div(Title(rcc.Commit.When.Format("01/02/2006 03:04:05 PM")), Text(humanize.Time(rcc.Commit.When))),
		),
		Details(Class("text-text mt-5"),
			Summary(Class("cursor-pointer"), Textf("%d changed files, %d additions(+), %d deletions(-)", rcc.Commit.Stats.Changed, rcc.Commit.Stats.Additions, rcc.Commit.Stats.Deletions)),
			Div(Class("p-3 bg-base rounded"),
				Map(rcc.Commit.Files, func(file git.CommitFile) Node {
					return A(Class("block underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href("#"+file.Path()), Text(file.Path()))
				}),
			),
		),
		Map(rcc.Commit.Files, func(file git.CommitFile) Node {
			return Group([]Node{
				Div(Class("text-text mt-5"), ID(file.Path()),
					Span(Class("text-text/80"), Title(file.Action), Text(string(file.Action[0]))),
					Text(" "),
					If(file.From.Path != "",
						A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/tree/%s/%s", rcc.RepoHeaderComponentContext.Name, file.From.Commit, file.From.Path)), Text(file.From.Path)),
					),
					If(file.From.Path != "" && file.To.Path != "", Text(" → ")),
					If(file.To.Path != "",
						A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/tree/%s/%s", rcc.RepoHeaderComponentContext.Name, file.To.Commit, file.To.Path)), Text(file.To.Path)),
					),
				),
				Div(Class("code"),
					Raw(file.Patch),
				),
			})
		}),
	}...)
}
