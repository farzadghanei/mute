#!/bin/env make -f

SHELL = /bin/sh
makefile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
makefile_dir := $(dir $(makefile_path))

# installation
DESTDIR ?=
prefix ?= /usr/local
exec_prefix ?= $(prefix)
bindir ?= $(exec_prefix)/bin

# use Make's builtin variable to call 'install'
INSTALL ?= install
INSTALL_PROGRAM ?= $(INSTALL)
INSTALL_DATA ?= $(INSTALL -m 644)

# packaging
PBUILDER_COMPONENTS ?= "main universe"
PBUILDER_RC ?= $(makefile_dir)/packaging/pbuilderrc

export ARCH ?= amd64
export DIST ?= xenial

# command aliases
cowbuilder = env DISTRIBUTION=$(DIST) ARCH=$(ARCH) BASEPATH=/var/cache/pbuilder/base-$(DIST)-$(ARCH).cow cowbuilder


mute:
	go build cmd/mute.go


build: mute


test:
	go test github.com/farzadghanei/mute


install: build
	$(INSTALL_PROGRAM) -d $(DESTDIR)$(bindir)
	$(INSTALL_PROGRAM) mute $(DESTDIR)$(bindir)

uninstall:
	rm $(DESTDIR)$(bindir)/mute

clean:
	rm -f mute
	go clean || true


distclean: clean

# override prefix so .deb package installs binaries to /usr/bin instead of /usr/local/bin
pkg-deb: export prefix = /usr
# requires a cowbuilder environment. see pkg-deb-setup
pkg-deb:
	(test ! -e debian && echo "no debian directory exists! creating one ..." && /bin/true) || (echo "debian directory exists. Remove to continue. aborting!" && /bin/false)
	# @TODO: find the package version
	tar --exclude-vcs -zcf ../mute_0.1.0.orig.tar.gz .
	cp -r packaging/debian debian
	DIST=$(DIST) ARCH=$(ARCH) BUILDER=cowbuilder GIT_PBUILDER_OPTIONS="--configfile=$(PBUILDER_RC)" git-pbuilder

# required:
# sudo apt-get install sudo build-essential git-pbuilder devscripts ubuntu-dev-tools
pkg-deb-setup:
	echo "creating a git-pbuilder environment with latest go version ..."
	DIST=$(DIST) ARCH=$(ARCH) git-pbuilder create --components=$(PBUILDER_COMPONENTS) --extrapackages="cowdancer" --configfile=$(PBUILDER_RC)
	echo "apt-get update; apt-get install -yq software-properties-common;" | sudo $(cowbuilder) --login --save-after-login
	echo "add-apt-repository ppa:longsleep/golang-backports; apt-get update;" | sudo $(cowbuilder) --login --save-after-login

pkg-clean:
	rm -rf debian


.DEFAULT_GOAL := build
.PHONY: test build install pkg-deb pkg-clean pkg-deb-setup
