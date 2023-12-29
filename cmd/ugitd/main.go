package main

import (
	"fmt"
	"os"
)

func main() {
	args, err := parseArgs(os.Args[1:])
	if err != nil {
		panic(err)
	}
	fmt.Println(args)
}
