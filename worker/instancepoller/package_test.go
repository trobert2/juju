// Copyright 2014 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the AGPLv3, see LICENCE file for details.

package instancepoller

import (
	"runtime"
	stdtesting "testing"

	coretesting "github.com/juju/juju/testing"
)

func TestAll(t *stdtesting.T) {
	if runtime.GOOS != "windows" {
		coretesting.MgoTestPackage(t)
	}
}
