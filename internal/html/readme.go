package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ReadmeComponentContext struct {
	Markdown string
}

func readmeComponent(rcc ReadmeComponentContext) Node {
	if rcc.Markdown == "" {
		return nil
	}
	return Div(Class("bg-base dark:bg-base/50 p-5 mt-5 rounded markdown"), Raw(rcc.Markdown))
}
