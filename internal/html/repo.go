package html

import (
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type RepoHeaderComponentContext struct {
	Name        string
	Ref         string
	Description string
	CloneURL    string
	Tags        []string
}

func repoHeaderComponent(rhcc RepoHeaderComponentContext) Node {
	return Group([]Node{
		Div(Class("mb-1 text-text"),
			A(Class("text-lg underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href("/"+rhcc.Name), Text(rhcc.Name)),
			If(rhcc.Ref != "", Group([]Node{
				Text(" "),
				A(Class("text-text/80 text-sm underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/tree/%s/", rhcc.Name, rhcc.Ref)), Text("@"+rhcc.Ref)),
			})),
			Text(" - "),
			A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/refs", rhcc.Name)), Text("refs")),
			Text(" - "),
			A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/log/%s", rhcc.Name, rhcc.Ref)), Text("log")),
			Text(" - "),
			Form(Class("inline-block"), Action(fmt.Sprintf("/%s/search", rhcc.Name)), Method("get"),
				Input(Class("rounded p-1 bg-mantle focus:border-lavender focus:outline-none focus:ring-0"), ID("search"), Type("text"), Name("q"), Placeholder("search")),
			),
			Text(" - "),
			Pre(Class("text-text inline select-all bg-base dark:bg-base/50 p-1 rounded"), Textf("%s/%s.git", rhcc.CloneURL, rhcc.Name)),
		),
		Div(Class("text-subtext0 mb-1"),
			Map(rhcc.Tags, func(tag string) Node {
				return Span(Class("rounded border-rosewater border-solid border pb-0.5 px-1 mr-1 mb-1 inline-block"), Text(tag))
			}),
		),
		Div(Class("text-text/80 mb-1"), Text(rhcc.Description)),
	})
}
