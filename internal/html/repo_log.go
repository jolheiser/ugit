package html

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"go.jolheiser.com/ugit/internal/git"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type RepoLogContext struct {
	BaseContext
	RepoHeaderComponentContext
	Commits []git.Commit
}

func RepoLogTemplate(rlc RepoLogContext) Node {
	return base(rlc.BaseContext, []Node{
		repoHeaderComponent(rlc.RepoHeaderComponentContext),
		Div(Class("grid sm:grid-cols-8 gap-1 text-text mt-5"),
			Map(rlc.Commits, func(commit git.Commit) Node {
				return Group([]Node{
					Div(Class("sm:col-span-5"),
						Div(
							A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/commit/%s", rlc.RepoHeaderComponentContext.Name, commit.SHA)), Text(commit.Short())),
						),
						Div(Class("whitespace-pre"),
							If(commit.Details() != "",
								Details(
									Summary(Class("cursor-pointer"), Text(commit.Summary())),
									Div(Class("p-3 bg-base rounded"), Text(commit.Details())),
								),
							),
							If(commit.Details() == "",
								Text(commit.Message),
							),
						),
					),
					Div(Class("sm:col-span-3 mb-4"),
						Div(
							Text(commit.Author),
							Text(" "),
							A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("mailto:%s", commit.Email)), Textf("<%s>", commit.Email)),
						),
						Div(Title(commit.When.Format("01/02/2006 03:04:05 PM")), Text(humanize.Time(commit.When))),
					),
				})
			}),
		),
	}...)
}
