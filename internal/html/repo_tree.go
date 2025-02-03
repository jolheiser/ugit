package html

import (
	"fmt"

	"go.jolheiser.com/ugit/internal/git"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type RepoTreeContext struct {
	BaseContext
	RepoHeaderComponentContext
	RepoBreadcrumbComponentContext
	RepoTreeComponentContext
	ReadmeComponentContext
	Description string
}

type RepoTreeComponentContext struct {
	Repo string
	Ref  string
	Tree []git.FileInfo
	Back string
}

func slashDir(name string, isDir bool) string {
	if isDir {
		return name + "/"
	}
	return name
}

func repoTreeComponent(rtcc RepoTreeComponentContext) Node {
	return Div(Class("grid grid-cols-3 sm:grid-cols-8 text-text py-5 rounded px-5 gap-x-3 gap-y-1 bg-base dark:bg-base/50"),
		If(rtcc.Back != "", Group([]Node{
			Div(Class("col-span-2")),
			Div(Class("sm:col-span-6"),
				A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/tree/%s/%s", rtcc.Repo, rtcc.Ref, rtcc.Back)), Text("..")),
			),
		})),
		Map(rtcc.Tree, func(fi git.FileInfo) Node {
			return Group([]Node{
				Div(Class("sm:col-span-1 break-keep"), Text(fi.Mode)),
				Div(Class("sm:col-span-1 text-right"), Text(fi.Size)),
				Div(Class("sm:col-span-6 overflow-hidden text-ellipsis"),
					A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/tree/%s/%s", rtcc.Repo, rtcc.Ref, fi.Path)), Text(slashDir(fi.Name(), fi.IsDir))),
				),
			})
		}),
	)
}

func RepoTreeTemplate(rtc RepoTreeContext) Node {
	return base(rtc.BaseContext, []Node{
		repoHeaderComponent(rtc.RepoHeaderComponentContext),
		repoBreadcrumbComponent(rtc.RepoBreadcrumbComponentContext),
		repoTreeComponent(rtc.RepoTreeComponentContext),
		readmeComponent(rtc.ReadmeComponentContext),
	}...)
}
