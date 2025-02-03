package html

import (
	"fmt"

	"go.jolheiser.com/ugit/internal/git"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type RepoRefsContext struct {
	BaseContext
	RepoHeaderComponentContext
	Branches []string
	Tags     []git.Tag
}

func RepoRefsTemplate(rrc RepoRefsContext) Node {
	return base(rrc.BaseContext, []Node{
		repoHeaderComponent(rrc.RepoHeaderComponentContext),
		If(len(rrc.Branches) > 0, Group([]Node{
			H3(Class("text-text text-lg mt-5"), Text("Branches")),
			Div(Class("text-text grid grid-cols-4 sm:grid-cols-8"),
				Map(rrc.Branches, func(branch string) Node {
					return Group([]Node{
						Div(Class("col-span-2 sm:col-span-1 font-bold"), Text(branch)),
						Div(Class("col-span-2 sm:col-span-7"),
							A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/tree/%s/", rrc.RepoHeaderComponentContext.Name, branch)), Text("tree")),
							Text(" "),
							A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/log/%s", rrc.RepoHeaderComponentContext.Name, branch)), Text("log")),
						),
					})
				}),
			),
		})),
		If(len(rrc.Tags) > 0, Group([]Node{
			H3(Class("text-text text-lg mt-5"), Text("Tags")),
			Div(Class("text-text grid grid-cols-8"),
				Map(rrc.Tags, func(tag git.Tag) Node {
					return Group([]Node{
						Div(Class("col-span-1 font-bold"), Text(tag.Name)),
						Div(Class("col-span-7"),
							A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/tree/%s/", rrc.RepoHeaderComponentContext.Name, tag.Name)), Text("tree")),
							Text(" "),
							A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/log/%s", rrc.RepoHeaderComponentContext.Name, tag.Name)), Text("log")),
						),
						If(tag.Signature != "",
							Details(Class("col-span-8 whitespace-pre"),
								Summary(Class("cursor-pointer"), Text("Signature")),
								Code(Text(tag.Signature)),
							),
						),
						If(tag.Annotation != "",
							Div(Class("col-span-8 mb-3"), Text(tag.Annotation)),
						),
					})
				}),
			),
		})),
	}...)
}
