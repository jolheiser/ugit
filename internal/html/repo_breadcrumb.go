package html

import (
	"fmt"
	"path"
	"strings"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type RepoBreadcrumbComponentContext struct {
	Repo string
	Ref  string
	Path string
}

type breadcrumb struct {
	label string
	href  string
	end   bool
}

func (r RepoBreadcrumbComponentContext) crumbs() []breadcrumb {
	parts := strings.Split(r.Path, "/")
	breadcrumbs := []breadcrumb{
		{
			label: r.Repo,
			href:  fmt.Sprintf("/%s/tree/%s/", r.Repo, r.Ref),
		},
	}
	for idx, part := range parts {
		breadcrumbs = append(breadcrumbs, breadcrumb{
			label: part,
			href:  path.Join(breadcrumbs[idx].href, part),
		})
	}
	breadcrumbs[len(breadcrumbs)-1].end = true
	return breadcrumbs
}

func repoBreadcrumbComponent(rbcc RepoBreadcrumbComponentContext) Node {
	if rbcc.Path == "" {
		return nil
	}
	return Div(Class("inline-block text-text"),
		Map(rbcc.crumbs(), func(crumb breadcrumb) Node {
			if crumb.end {
				return Span(Text(crumb.label))
			}
			return Group([]Node{
				A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(crumb.href), Text(crumb.label)),
				Text(" / "),
			})
		}),
	)
}
