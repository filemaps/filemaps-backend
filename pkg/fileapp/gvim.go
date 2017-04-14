// Copyright (C) 2017 File Maps Backend Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package fileapp

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
)

const (
	gvimBin = "/usr/bin/gvim"
)

func init() {
	// register file app if binary exists
	if _, err := os.Stat(gvimBin); err == nil {
		register(NewGVim())
	}
}

type GVim struct {
}

func NewGVim() *GVim {
	return &GVim{}
}

func (a *GVim) getName() string {
	return "GVim"
}

func (a *GVim) open(path string) int {
	log.WithFields(log.Fields{
		"path": path,
	}).Info("GVim: open")

	out, err := exec.Command(gvimBin, "-p", "--remote-tab-silent", path).Output()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("GVim open error")
		return -1
	}
	log.WithFields(log.Fields{
		"out": out,
	}).Info("GVim")
	return 0
}