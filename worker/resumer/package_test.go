// Copyright 2014 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package resumer_test

import (
	"runtime"
	stdtesting "testing"

	"github.com/juju/juju/testing"
)

func Test(t *stdtesting.T) {
	if runtime.GOOS != "windows" {
		testing.MgoTestPackage(t)
	}
}
