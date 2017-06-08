package server

import "io/ioutil"

// todo: make this object oriented

const defaultSSHWrapper = "/tmp/git-ssh.sh"

// Create git ssh wrapper
func GenerateSSHWrapper() error {
	d1 := []byte("#!/bin/sh\nif [ -z \"$PKEY\" ]; then\n# if PKEY is not specified, run ssh using default keyfile\nssh -oStrictHostKeyChecking=no \"$@\"\nelse\nssh -oStrictHostKeyChecking=no -i \"$PKEY\" \"$@\"\nfi")
	return ioutil.WriteFile(defaultSSHWrapper, d1, 0755)
}
