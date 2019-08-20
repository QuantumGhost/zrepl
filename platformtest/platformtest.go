package platformtest

import (
	"context"
	"log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Context struct {
	context.Context
	RootDataset string
}

var _ assert.TestingT = (*Context)(nil)
var _ require.TestingT = (*Context)(nil)

func (c *Context) Errorf(format string, args ...interface{}) {
	// getLog(c).Printf(format, args...)
	log.Printf(format, args...)
}

func (c *Context) FailNow() {
	panic(nil)
}
