package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/fatih/color"
	"github.com/zrepl/zrepl/logger"
	"github.com/zrepl/zrepl/platformtest"
	"github.com/zrepl/zrepl/platformtest/tests"
)

func main() {

	root := flag.String("root", "", "empty root filesystem under which we conduct the platform test")
	flag.Parse()
	if *root == "" {
		panic(*root)
	}

	ctx := &platformtest.Context{
		platformtest.WithLogger(context.Background(), logger.NewStderrDebugLogger()),
		*root,
	}

	bold := color.New(color.Bold)
	for _, c := range tests.Cases {
		bold.Printf("BEGIN TEST CASE %s\n", c)
		c(ctx)
		bold.Printf("DONE  TEST CASE %s\n", c)
		fmt.Println()
	}

}
