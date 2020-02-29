#!/bin/env make -f
# license: MIT, see LICENSE for details.

SHELL = /bin/sh
makefile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
makefile_dir := $(dir $(makefile_path))
MUTE_LATEST_TAG := $(shell git tag --list | grep --only-matching --line-regexp --perl-regexp '\d+\.\d+\.\d+' | uniq | sort -V | tail -n 1)

# build
OS ?= linux
ARCH ?= amd64
DIST ?= xenial
LDFLAGS ?= "-s"  # by default create a leaner binary
GOARCH ?= amd64

ifeq ($(ARCH), amd64)
    GOARCH = amd64
else ifeq ($(ARCH), i368)
    GOARCH = 386
endif

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
PKG_TGZ_NAME = mute-$(MUTE_LATEST_TAG)-$(OS)-$(ARCH).tar.gz
PBUILDER_COMPONENTS ?= "main universe"
PBUILDER_RC ?= $(makefile_dir)/packaging/pbuilderrc

# find Debian package version from the changelog file. latest version
# should be at the top, first matching 'mute (0.1.0) ...' and sed clears chars not in version
MUTE_DEB_VERSION := $(shell grep --only-matching --max-count 1 --perl-regexp "^\s*mute\s+\(.+\)\s*" packaging/debian/changelog | sed 's/[^0-9.]//g')


# command aliases
cowbuilder = env DISTRIBUTION=$(DIST) ARCH=$(ARCH) BASEPATH=/var/cache/pbuilder/base-$(DIST)-$(ARCH).cow cowbuilder


mute:
	GOOS=$(OS) GOARCH=$(GOARCH) go build -ldflags $(LDFLAGS) cmd/mute.go


build: mute


test:
	go test github.com/farzadghanei/mute

test-build: build
	./mute fixtures/xecho -c 3 > /dev/null; (test "$$?" -eq 3 || false)
	./mute fixtures/xecho -c 1 'not muted' | grep -q 'not muted'
	output=$$(env MUTE_EXIT_CODES=1 ./mute fixtures/xecho -c 1 'muted'); test -z "$$output"
	env MUTE_EXIT_CODDE=1 ./mute fixtures/xecho -c 2 'not muted' | grep -q 'not muted'
	output=$$(env MUTE_STDOUT_PATTERN='mute.+' ./mute fixtures/xecho 'will be muted.'); test -z "$$output"
	env MUTE_STDOUT_PATTERN='nottoday' ./mute fixtures/xecho 'not muted' | grep -q 'not muted'


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
	tar --exclude-vcs -zcf ../mute_$(MUTE_DEB_VERSION).orig.tar.gz .
	cp -r packaging/debian debian
	DIST=$(DIST) ARCH=$(ARCH) BUILDER=cowbuilder GIT_PBUILDER_OPTIONS="--configfile=$(PBUILDER_RC)" git-pbuilder

# required:
# sudo apt-get install sudo build-essential git-pbuilder devscripts ubuntu-dev-tools
pkg-deb-setup:
	echo "creating a git-pbuilder environment with latest go version ..."
	DIST=$(DIST) ARCH=$(ARCH) git-pbuilder create --components=$(PBUILDER_COMPONENTS) --extrapackages="cowdancer" --configfile=$(PBUILDER_RC)
	echo "apt-get update; apt-get install -yq software-properties-common;" | sudo $(cowbuilder) --login --save-after-login
	echo "add-apt-repository ppa:longsleep/golang-backports; apt-get update;" | sudo $(cowbuilder) --login --save-after-login

pkg-tgz: build
	tar --create --gzip --exclude-vcs --exclude=docs/man/*.rst --file $(PKG_TGZ_NAME) mute README.rst LICENSE docs/man/mute.1

pkg-clean:
	rm -rf debian
	rm -f $(PKG_TGZ_NAME)

# required: python docutils
docs:
	rst2man.py --input-encoding=utf8 --output-encoding=utf8 --strict docs/man/mute.rst docs/man/mute.1

.DEFAULT_GOAL := build
.PHONY: test build test-build install pkg-deb pkg-clean pkg-deb-setup pkg-tgz docs
