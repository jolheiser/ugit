package html

import (
	"fmt"

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

func lastCommitTime(repo *git.Repo, human bool) string {
	c, err := repo.LastCommit()
	if err != nil {
		return ""
	}
	if human {
		return humanize.Time(c.When)
	}
	return c.When.Format("01/02/2006 03:04:05 PM")
}

func lastCommit(repo *git.Repo) *git.Commit {
	c, err := repo.LastCommit()
	if err != nil {
		return nil
	}
	return &c
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
			Div(Class("grid sm:grid-cols-10 gap-2 mt-5"),
				Map(ic.Repos, func(repo *git.Repo) Node {
					commit := lastCommit(repo)
					return Group([]Node{
						Div(Class("sm:col-span-2 text-blue dark:text-lavender"),
							A(Class("underline decoration-blue/50 dark:decoration-lavender/50 decoration-dashed hover:decoration-solid"), Href("/"+repo.Name()), Text(repo.Name())),
						),
						Div(Class("sm:col-span-3 text-subtext0"), Text(repo.Meta.Description)),
						Div(Class("sm:col-span-3 text-subtext0"),
							If(commit != nil,
								Div(Title(commit.Message),
									A(Class("underline text-blue dark:text-lavender decoration-blue/50 dark:decoration-lavender/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/commit/%s", repo.Name(), commit.SHA)), Text(commit.Short())),
									Text(": "+commit.Summary()),
								),
							),
						),
						Div(Class("sm:col-span-1 text-subtext0"),
							Map(repo.Meta.Tags, func(tag string) Node {
								return A(Class("rounded border-rosewater border-solid border pb-0.5 px-1 mr-1 mb-1 inline-block"), Href("?tag="+tag), Text(tag))
							}),
						),
						Div(Class("sm:col-span-1 text-text/80 mb-4 sm:mb-0"), Title(lastCommitTime(repo, false)), Text(lastCommitTime(repo, true))),
					})
				}),
			),
		),
	}...)
}
