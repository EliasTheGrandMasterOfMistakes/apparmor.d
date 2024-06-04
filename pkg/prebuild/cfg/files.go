// apparmor.d - Full set of apparmor profiles
// Copyright (C) 2021-2024 Alexandre Pujol <alexandre@pujol.io>
// SPDX-License-Identifier: GPL-2.0-only

package cfg

import (
	"fmt"
	"os"
	"strings"

	"github.com/roddhjav/apparmor.d/pkg/paths"
	"github.com/roddhjav/apparmor.d/pkg/util"
)

// Default content of debian/apparmor.d.hide. Whonix has special addition.
var Hide = `# This file is generated by "make", all edit will be lost.

/etc/apparmor.d/usr.bin.firefox
/etc/apparmor.d/usr.sbin.cups-browsed
/etc/apparmor.d/usr.sbin.cupsd
/etc/apparmor.d/usr.sbin.rsyslogd
`

type Flagger struct{}

func (f Flagger) Read(name string) map[string][]string {
	res := map[string][]string{}
	path := FlagDir.Join(name + ".flags")
	if !path.Exist() {
		return res
	}

	lines := util.MustReadFileAsLines(path)
	for _, line := range lines {
		manifest := strings.Split(line, " ")
		profile := manifest[0]
		flags := []string{}
		if len(manifest) > 1 {
			flags = strings.Split(manifest[1], ",")
		}
		res[profile] = flags
	}
	return res
}

type Ignorer struct{}

func (i Ignorer) Read(name string) []string {
	path := IgnoreDir.Join(name + ".ignore")
	if !path.Exist() {
		return []string{}
	}
	return util.MustReadFileAsLines(path)
}

type Overwriter bool

// Overwrite upstream profile: disable upstream & rename ours
func (o Overwriter) Apply() error {
	const ext = ".apparmor.d"
	disableDir := RootApparmord.Join("disable")
	if err := disableDir.Mkdir(); err != nil {
		return err
	}

	path := DistDir.Join("overwrite")
	if !path.Exist() {
		return fmt.Errorf("%s not found", path)
	}
	for _, name := range util.MustReadFileAsLines(path) {
		origin := RootApparmord.Join(name)
		dest := RootApparmord.Join(name + ext)
		if err := origin.Rename(dest); err != nil {
			return err
		}
		originRel, err := origin.RelFrom(dest)
		if err != nil {
			return err
		}
		if err := os.Symlink(originRel.String(), disableDir.Join(name).String()); err != nil {
			return err
		}
	}
	return nil
}

type DebianHider struct {
	path *paths.Path
}

// Initialize the file with content from Hide
func (d DebianHider) Init() error {
	return d.path.WriteFile([]byte(Hide))
}

// Initialize the file with content from Hide
func (d DebianHider) Clean() error {
	return d.path.WriteFile([]byte("# This file is generated by \"make\", all edit will be lost.\n"))
}
