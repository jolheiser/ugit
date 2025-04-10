package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type BaseContext struct {
	Title       string
	Description string
}

func base(bc BaseContext, children ...Node) Node {
	return HTML5(HTML5Props{
		Title:       bc.Title,
		Description: bc.Description,
		Head: []Node{
			Link(Rel("icon"), Href("/_/favicon.svg")),
			Link(Rel("stylesheet"), Href("/_/tailwind.css")),
			ogp("title", bc.Title),
			ogp("description", bc.Description),
			Meta(Name("forge"), Content("ugit")),
			Meta(Name("keywords"), Content("git,forge,ugit")),
		},
		Body: []Node{
			Class("latte dark:mocha bg-base/50 dark:bg-base/95 max-w-7xl mx-5 sm:mx-auto my-10"),
			H2(Class("text-text text-xl mb-3"),
				A(Class("text-text text-xl mb-3"), Href("/"), Text("Home")),
			),
			Group(children),
		},
	})
}

func ogp(property, content string) Node {
	return El("meta", Attr("property", "og:"+property), Attr("content", content))
}
