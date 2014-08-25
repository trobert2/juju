// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package version

import (
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils/set"
	gc "launchpad.net/gocheck"
)

type macOSXVersionSuite struct{}

var _ = gc.Suite(&macOSXVersionSuite{})

func (s *macOSXVersionSuite) TestOSVersion(c *gc.C) {
	knownSeries := set.Strings{}
	for _, series := range macOSXSeries {
		knownSeries.Add(series)
	}
	version, err := osVersion()
	c.Assert(err, gc.IsNil)
	c.Check(version, jc.Satisfies, knownSeries.Contains)
}
