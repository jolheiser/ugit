package main

import (
	"flag"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffyaml"
)

type args struct {
	db string
}

func parseArgs(args []string) (a args, e error) {
	fs := flag.NewFlagSet("ugitd", flag.ContinueOnError)
	fs.String("config", "ugit.yaml", "Path to config file")
	return a, ff.Parse(fs, args,
		ff.WithEnvVarPrefix("UGIT"),
		ff.WithConfigFileFlag("config"),
		ff.WithAllowMissingConfigFile(true),
		ff.WithConfigFileParser(ffyaml.Parser),
	)
}
