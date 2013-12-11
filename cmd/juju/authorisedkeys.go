// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package main

import (
	"launchpad.net/juju-core/cmd"
)

var authKeysDoc = `
"juju authorized-keys" is used to manage the ssh keys allowed to log on to
nodes in the Juju environment.

`

type AuthorisedKeysCommand struct {
	*cmd.SuperCommand
}

func NewAuthorisedKeysCommand() cmd.Command {
	sshkeyscmd := &AuthorisedKeysCommand{
		SuperCommand: cmd.NewSuperCommand(cmd.SuperCommandParams{
			Name:    "authorised-keys",
			Doc:     authKeysDoc,
			Purpose: "manage authorised ssh keys",
		}),
	}
	sshkeyscmd.Register(&ListKeysCommand{})
	return sshkeyscmd
}
