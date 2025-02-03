package html

import (
	"github.com/dustin/go-humanize"
	"go.jolheiser.com/ugit/assets"
	"go.jolheiser.com/ugit/internal/git"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type IndexContext struct {
	BaseContext
	Profile IndexProfile
	Repos   []*git.Repo
}

type IndexProfile struct {
	Username string
	Email    string
	Links    []IndexLink
}

type IndexLink struct {
	Name string
	URL  string
}

func lastCommit(repo *git.Repo, human bool) string {
	c, err := repo.LastCommit()
	if err != nil {
		return ""
	}
	if human {
		return humanize.Time(c.When)
	}
	return c.When.Format("01/02/2006 03:04:05 PM")
}

func IndexTemplate(ic IndexContext) Node {
	return base(ic.BaseContext, []Node{
		Header(
			H1(Class("text-text text-xl font-bold"), Text(ic.Title)),
			H2(Class("text-subtext1 text-lg"), Text(ic.Description)),
		),
		Main(Class("mt-5"),
			Div(Class("grid grid-cols-1 sm:grid-cols-8"),
				If(ic.Profile.Username != "",
					Div(Class("text-mauve"), Text("@"+ic.Profile.Username)),
				),
				If(ic.Profile.Email != "", Group([]Node{
					Div(Class("text-mauve col-span-2"),
						Div(Class("w-5 h-5 stroke-mauve inline-block mr-1 align-middle"), Raw(string(assets.EmailIcon))),
						A(Class("underline decoration-mauve/50 decoration-dashed hover:decoration-solid"), Href("mailto:"+ic.Profile.Email), Text(ic.Profile.Email)),
					),
				}),
				),
			),
			Div(Class("grid grid-cols-1 sm:grid-cols-8"),
				Map(ic.Profile.Links, func(link IndexLink) Node {
					return Div(Class("text-mauve"),
						Div(Class("w-5 h-5 stroke-mauve inline-block mr-1 align-middle"),
							Raw(string(assets.LinkIcon)),
						),
						A(Class("underline decoration-mauve/50 decoration-dashed hover:decoration-solid"), Rel("me"), Href(link.URL), Text(link.Name)),
					)
				}),
			),
			Div(Class("grid sm:grid-cols-8 gap-2 mt-5"),
				Map(ic.Repos, func(repo *git.Repo) Node {
					return Group([]Node{
						Div(Class("sm:col-span-2 text-blue dark:text-lavender"),
							A(Class("underline decoration-blue/50 dark:decoration-lavender/50 decoration-dashed hover:decoration-solid"), Href("/"+repo.Name()), Text(repo.Name())),
						),
						Div(Class("sm:col-span-4 text-subtext0"), Text(repo.Meta.Description)),
						Div(Class("sm:col-span-1 text-subtext0"),
							Map(repo.Meta.Tags, func(tag string) Node {
								return A(Class("rounded border-rosewater border-solid border pb-0.5 px-1 mr-1 mb-1 inline-block"), Href("?tag="+tag), Text(tag))
							}),
						),
						Div(Class("sm:col-span-1 text-text/80 mb-4 sm:mb-0"), Title(lastCommit(repo, false)), Text(lastCommit(repo, true))),
					})
				}),
			),
		),
	}...)
}
