package html

import (
	_ "embed"
	"fmt"

	"go.jolheiser.com/ugit/internal/git"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SearchContext struct {
	BaseContext
	RepoHeaderComponentContext
	Results []git.GrepResult
}

func (s SearchContext) DedupeResults() [][]git.GrepResult {
	var (
		results     [][]git.GrepResult
		currentFile string
	)
	var idx int
	for _, result := range s.Results {
		if result.File == currentFile {
			results[idx-1] = append(results[idx-1], result)
			continue
		}
		results = append(results, []git.GrepResult{result})
		currentFile = result.File
		idx++
	}

	return results
}

//go:embed repo_search.js
var repoSearchJS string

func repoSearchResult(repo, ref string, results []git.GrepResult) Node {
	return Group([]Node{
		Div(Class("text-text mt-5"),
			A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/tree/%s/%s#L%d", repo, ref, results[0].File, results[0].Line)), Text(results[0].File)),
		),
		Div(Class("code"),
			Raw(results[0].Content),
		),
		If(len(results) > 1,
			Details(Class("text-text cursor-pointer"),
				Summary(Textf("%d more", len(results[1:]))),
				Map(results[1:], func(result git.GrepResult) Node {
					return Group([]Node{
						Div(Class("text-text mt-5 ml-5"),
							A(Class("underline decoration-text/50 decoration-dashed hover:decoration-solid"), Href(fmt.Sprintf("/%s/tree/%s/%s#L%d", repo, ref, result.File, result.Line)), Text(results[0].File)),
						),
						Div(Class("code ml-5"),
							Raw(result.Content),
						),
					})
				}),
			),
		),
	})
}

func RepoSearchTemplate(sc SearchContext) Node {
	dedupeResults := sc.DedupeResults()
	return base(sc.BaseContext, []Node{
		repoHeaderComponent(sc.RepoHeaderComponentContext),
		Map(dedupeResults, func(results []git.GrepResult) Node {
			return repoSearchResult(sc.RepoHeaderComponentContext.Name, sc.RepoHeaderComponentContext.Ref, results)
		}),
		If(len(dedupeResults) == 0,
			P(Class("text-text mt-5 text-lg"), Text("No results")),
		),
		Script(Text(repoSearchJS)),
	}...)
}
