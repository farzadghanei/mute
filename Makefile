#!/bin/env make -f

# license: MIT, see LICENSE for details.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy of this software
# and associated documentation files (the "Software"), to deal in the Software without restriction,
# including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
# subject to the following conditions:
# The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
# WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
#
SHELL = /bin/sh
makefile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
makefile_dir := $(dir $(makefile_path))
MUTE_VERSION := $(shell grep --perl-regex '^\s*const\s+Version\s+string' conf.go | grep --only-matching --perl-regexp '[\d\.]+')
TIMESTAMP_MINUTE := $(shell date -u +%Y%m%d%H%M)

# build
OS ?= linux
ARCH ?= amd64
DIST ?= trixie  # go 1.24 is available in trixie
GOLDFLAGS ?= "-s"  # by default create a leaner binary
GOARCH ?= amd64
GOBUILDFLAGS ?=  # additional build arguments

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
sharedir ?= $(prefix)/share
mandir ?= $(sharedir)/man/man1
docsdir ?= $(sharedir)/doc/mute

# use Make's builtin variable to call 'install'
INSTALL ?= install
INSTALL_PROGRAM ?= $(INSTALL)
INSTALL_DATA ?= $(INSTALL -m 644)

# store current git branch, so after building a package in a temp local branch
# build process can return to the original branch
GIT_CURRENT_BRANCH := $(shell git branch --show-current)

# packaging
SHA256SUM ?= sha256sum -b
PKG_DIST_DIR ?= $(abspath $(makefile_dir)/..)
PKG_TGZ_NAME = mute-$(MUTE_VERSION)-$(OS)-$(ARCH).tar.gz
PKG_TGZ_PATH = $(PKG_DIST_DIR)/$(PKG_TGZ_NAME)
PKG_CHECKSUM_NAME = mute-$(MUTE_VERSION)-SHA256SUMS

PBUILDER_COMPONENTS ?= "main universe"
PBUILDER_RC ?= $(makefile_dir)build/package/pbuilderrc
PBUILDER_HOOKS_DIR ?= $(makefile_dir)build/package/pbuilder-hooks

RPM_DEV_TREE ?= $(HOME)/rpmbuild

# find Debian package version from the changelog file. latest version
# should be at the top, first matching 'mute (0.1.0-1) ...' and sed clears chars not in version
MUTE_DEB_VERSION := $(shell grep --only-matching --max-count 1 --perl-regexp "^\s*mute\s+\(.+\)\s*" build/package/debian/changelog | sed 's/[^0-9.-]//g')
MUTE_DEB_UPSTREAM_VERSION := $(shell echo $(MUTE_DEB_VERSION) | grep --only-matching --perl-regexp '^[0-9.]+')
MUTE_DEB_UPSTREAM_TARBAL_PATH := $(abspath $(makefile_dir)/..)
MUTE_DEB_UPSTREAM_TARBAL := $(MUTE_DEB_UPSTREAM_TARBAL_PATH)/mute_$(MUTE_DEB_UPSTREAM_VERSION).orig.tar.gz
DEB_BUILD_GIT_BRANCH := pkg-deb-$(MUTE_DEB_VERSION)-$(TIMESTAMP_MINUTE)

# find rpm version from the spec file. latest version
# should be in the top tags, first matching 'Version: 0.1.0' and sed clears chars not in version
MUTE_RPM_VERSION := $(shell grep --only-matching --max-count 1 --line-regexp --perl-regexp "\s*Version\:\s*.+\s*" build/package/mute.spec | sed 's/[^0-9.]//g')
RPM_DEV_SRC_TGZ = $(RPM_DEV_TREE)/SOURCES/mute-$(MUTE_RPM_VERSION).tar.gz
RPM_DEV_SPEC = $(RPM_DEV_TREE)/SPECS/mute-$(MUTE_RPM_VERSION).spec

# command aliases
cowbuilder = env DISTRIBUTION=$(DIST) ARCH=$(ARCH) BASEPATH=/var/cache/pbuilder/base-$(DIST)-$(ARCH).cow cowbuilder

# testing
TEST_SKIP_STATICCHECKS ?=

mute:
	GOOS=$(OS) GOARCH=$(GOARCH) go build -ldflags $(GOLDFLAGS) $(GOBUILDFLAGS) cmd/mute.go

build: mute

test:
	go test -v -race ./...
	# @TODO: add static checks
	# if test -z $(TEST_SKIP_STATICCHECKS); then ./scripts/staticchecks; fi

test-build: build
	./mute test/data/xecho -c 3 > /dev/null; (test "$$?" -eq 3 || false)
	./mute test/data/xecho -c 1 'not muted' | grep -q 'not muted'
	output=$$(env MUTE_EXIT_CODES=1 ./mute test/data/xecho -c 1 'muted'); test -z "$$output"
	env MUTE_EXIT_CODDE=1 ./mute test/data/xecho -c 2 'not muted' | grep -q 'not muted'
	output=$$(env MUTE_STDOUT_PATTERN='mute.+' ./mute test/data/xecho 'will be muted.'); test -z "$$output"
	env MUTE_STDOUT_PATTERN='nottoday' ./mute test/data/xecho 'not muted' | grep -q 'not muted'

install: build
	$(INSTALL_PROGRAM) -d $(DESTDIR)$(bindir)
	$(INSTALL_PROGRAM) mute $(DESTDIR)$(bindir)
	mkdir -p $(DESTDIR)$(mandir)
	cp docs/man/mute.1 $(DESTDIR)$(mandir)
	mkdir -p $(DESTDIR)$(docsdir)
	cp README.rst $(DESTDIR)$(docsdir)/README.rst

uninstall:
	rm $(DESTDIR)$(bindir)/mute
	rm -f $(DESTDIR)$(mandir)/mute.1
	rm -f $(DESTDIR)$(docsdir)/*
	rmdir $(DESTDIR)$(docsdir)

clean:
	rm -f mute
	go clean || true

distclean: clean

# override prefix so .deb package installs binaries to /usr/bin instead of /usr/local/bin
pkg-deb: export prefix = /usr
# requires a cowbuilder environment. see pkg-deb-setup
pkg-deb:
	git checkout -b $(DEB_BUILD_GIT_BRANCH)
	rm -f $(MUTE_DEB_UPSTREAM_TARBAL); tar --exclude-backups --exclude-vcs -zcf $(MUTE_DEB_UPSTREAM_TARBAL) .
	cp -r build/package/debian .; git add debian; git commit -m 'add debian dir for packaging v$(MUTE_DEB_VERSION)'
	gbp buildpackage --git-ignore-new --git-verbose --git-pbuilder \
			 --git-no-create-orig --git-tarball-dir=$(MUTE_DEB_UPSTREAM_TARBAL_PATH) \
			 --git-hooks \
			 --git-dist=$(DIST) --git-arch=$(ARCH) \
			 --git-ignore-new --git-ignore-branch \
			 --git-pbuilder-options='--configfile=$(PBUILDER_RC) --hookdir=$(PBUILDER_HOOKS_DIR) --buildresult=$(PKG_DIST_DIR)' \
			 -b -us -uc -sa
	git checkout $(GIT_CURRENT_BRANCH)
	git branch -D $(DEB_BUILD_GIT_BRANCH)

# required:
# sudo apt-get install build-essential debhelper pbuilder fakeroot cowbuilder git-buildpackage devscripts ubuntu-dev-tools
pkg-deb-setup:
	echo "creating a git-pbuilder environment with apt repositories to install new go versions ..."
	DIST=$(DIST) ARCH=$(ARCH) git-pbuilder create --components=$(PBUILDER_COMPONENTS) \
							--extrapackages="cowdancer curl wget" --configfile=$(PBUILDER_RC) \
							--hookdir=$(PBUILDER_HOOKS_DIR)

pkg-tgz: build
	tar --create --gzip --exclude-vcs --exclude=docs/man/*.rst --file $(PKG_TGZ_PATH) mute \
		README.rst LICENSE docs/man/mute.1

# override prefix so .rpm package installs binaries to /usr/bin instead of /usr/local/bin
pkg-rpm: export prefix = /usr
# requires golang compiler > 1.13, and rpmdevtools package
pkg-rpm:
	(go version | grep -q go1.2[4-9]) || (echo "please install Go lang tools > 1.24. aborting!" && /bin/false)
	mkdir -p $(RPM_DEV_TREE)/RPMS $(RPM_DEV_TREE)/SRPMS $(RPM_DEV_TREE)/SOURCES $(RPM_DEV_TREE)/SPECS
	rm -f $(RPM_DEV_SRC_TGZ)
	tar --exclude-vcs -zcf $(RPM_DEV_SRC_TGZ) .
	cp build/package/mute.spec $(RPM_DEV_SPEC)
	rpmbuild -bs $(RPM_DEV_SPEC)
	rpmbuild --rebuild $(RPM_DEV_TREE)/SRPMS/mute-$(MUTE_RPM_VERSION)*.src.rpm
	find $(RPM_DEV_TREE)/RPMS -type f -readable -name 'mute-$(MUTE_RPM_VERSION)*.rpm' -exec mv '{}' $(PKG_DIST_DIR) \;

pkg-clean:
	rm -rf debian
	rm -f $(PKG_TGZ_NAME)

pkg-checksum:
	if test -e $(PKG_TGZ_PATH); then cd $(PKG_DIST_DIR) \
	    && (sed -i '/$(PKG_TGZ_NAME)/d' $(PKG_CHECKSUM_NAME) || true) \
	    && $(SHA256SUM) $(PKG_TGZ_NAME) >> $(PKG_CHECKSUM_NAME); fi
	if test -e $(PKG_DIST_DIR); then cd $(PKG_DIST_DIR) \
	    && (sed -i '/mute_$(MUTE_DEB_VERSION).*deb/d' $(PKG_CHECKSUM_NAME) || true) \
	    && find . -maxdepth 1 -readable -type f -name 'mute_$(MUTE_DEB_VERSION)*.deb' \
	    -exec sha256sum '{}' \; >> $(PKG_CHECKSUM_NAME); fi
	if test -e $(PKG_DIST_DIR); then cd $(PKG_DIST_DIR) \
	    && (sed -i '/mute-$(MUTE_RPM_VERSION).*rpm/d' $(PKG_CHECKSUM_NAME) || true) \
	    && find . -maxdepth 1 -readable -type f -name 'mute-$(MUTE_RPM_VERSION)*.rpm' \
	    -exec sha256sum '{}' \; >> $(PKG_CHECKSUM_NAME); fi

# sync version from source code to other files (docs, packaging, etc.)
sync-version:
	@ # update first occurrence of version in man page source and regen the man page
	@ sed -i 's/^:Version:.*/:Version: $(MUTE_VERSION)/;t' docs/man/mute.rst && $(MAKE) docs
	@ # update first occurrence of version  in RPM spec
	@ sed -i 's/^Version:.*/Version: $(MUTE_VERSION)/;t' build/package/mute.spec
	@ (grep --max 1 --only --perl-regexp '^mute.+\(.+\).+' build/package/debian/changelog | grep -q -F $(MUTE_VERSION)) || \
	    echo -e "\e[33m*** NOTE:\e[m version $(MUTE_VERSION) maybe missing from Debian changelog"
	@ (grep --line-regexp '%changelog' -A 50 build/package/mute.spec | grep -q -F $(MUTE_VERSION)) || \
	    echo -e "\e[33m*** NOTE:\e[m version $(MUTE_VERSION) maybe missing from RPM changelog"

# required: python docutils. install python3-docutils
docs:
	rst2man --input-encoding=utf8 --output-encoding=utf8 --strict docs/man/mute.rst docs/man/mute.1

.DEFAULT_GOAL := build
.PHONY: test build test-build install pkg-deb pkg-clean pkg-deb-setup pkg-tgz pkg-checksum sync-version docs echo-vars
