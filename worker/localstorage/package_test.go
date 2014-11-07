// Copyright 2014 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package localstorage_test

import (
	"runtime"
	stdtesting "testing"

	gc "gopkg.in/check.v1"
)

func TestAll(t *stdtesting.T) {
	if runtime.GOOS != "windows" {
		gc.TestingT(t)
	}
}
