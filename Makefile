#!/bin/env make -f

SHELL = /bin/sh

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
export ARCH := amd64
export DIST := xenial
export AUTO_DEBSIGN := no
export DEBUILD_DPKG_BUILDPACKAGE_OPTS := "-us -uc -I -i"
export DEBUILD_LINTIAN_OPTS := "-i -I --show-overrides"
export USENETWORK := yes
export BUILD_HOME := /build

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
	DIST=$(DIST) ARCH=$(ARCH) BUILDER=cowbuilder git-pbuilder

# required:
# sudo apt-get install sudo build-essential git-pbuilder devscripts ubuntu-dev-tools
# set these options set on ~/.pbuildrc (@TODO: set the variables during the build process)
# USENETWORK=yes
# BUILD_HOME=$BUILDDIR
# COMPONENTS="main universe"
# EXTRAPACKAGES="cowdancer software-properties-common"
pkg-deb-setup:
	echo "creating a git-pbuilder environment with latest go version ..."
	DIST=$(DIST) ARCH=$(ARCH) git-pbuilder create
	echo "apt-get update; apt-get install -yq software-properties-common;" | sudo $(cowbuilder) --login --save-after-login
	echo "add-apt-repository ppa:longsleep/golang-backports; apt-get update; apt-get -yq install golang" | sudo $(cowbuilder) --login --save-after-login

pkg-clean:
	rm -rf debian


.DEFAULT_GOAL := build
.PHONY: test build install pkg-deb pkg-clean pkg-deb-setup
