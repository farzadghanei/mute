#!/bin/env make -f
# license: MIT, see LICENSE for details.

SHELL = /bin/sh
makefile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
makefile_dir := $(dir $(makefile_path))
MUTE_LATEST_TAG := $(shell git tag --list | grep --only-matching --line-regexp --perl-regexp '\d+\.\d+\.\d+' | uniq | sort -V | tail -n 1)
MUTE_VERSION := $(shell grep --perl-regex '^\s*const\s+Version\s+string' conf.go | grep --only-matching --perl-regexp '[\d\.]+')
TIMESTAMP_MINUTE := $(shell date -u +%Y%m%d%H%M)

# build
OS ?= linux
ARCH ?= amd64
DIST ?= xenial
GOLDFLAGS ?= "-s"  # by default create a leaner binary
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

# newer Git versions support "git branch --show-current", this is for compatibiliy with older versions (2.25.4)
GIT_CURRENT_BRANCH := $(shell git branch --contains=HEAD | grep --line-regexp '\* .*' --max-count=1 | sed 's/\* //')

# packaging
SHA256SUM ?= sha256sum -b
PKG_DIST_DIR ?= $(abspath $(makefile_dir)/..)
PKG_TGZ_NAME = mute-$(MUTE_VERSION)-$(OS)-$(ARCH).tar.gz
PKG_TGZ_PATH = $(PKG_DIST_DIR)/$(PKG_TGZ_NAME)
PKG_CHECKSUM_NAME = mute-$(MUTE_VERSION)-SHA256SUMS
PBUILDER_COMPONENTS ?= "main universe"
PBUILDER_RC ?= $(makefile_dir)packaging/pbuilderrc
PBUILDER_HOOKS_DIR ?= $(makefile_dir)packaging/pbuilder-hooks
RPM_DEV_TREE ?= $(HOME)/rpmbuild

# find Debian package version from the changelog file. latest version
# should be at the top, first matching 'mute (0.1.0-1) ...' and sed clears chars not in version
MUTE_DEB_VERSION := $(shell grep --only-matching --max-count 1 --perl-regexp "^\s*mute\s+\(.+\)\s*" packaging/debian/changelog | sed 's/[^0-9.-]//g')
MUTE_DEB_UPSTREAM_VERSION := $(shell echo $(MUTE_DEB_VERSION) | grep --only-matching --perl-regexp '^[0-9.]+')
MUTE_DEB_UPSTREAM_TARBAL_PATH := $(abspath $(makefile_dir)/..)
MUTE_DEB_UPSTREAM_TARBAL := $(MUTE_DEB_UPSTREAM_TARBAL_PATH)/mute_$(MUTE_DEB_UPSTREAM_VERSION).orig.tar.gz
DEB_BUILD_GIT_BRANCH := pkg-deb-$(MUTE_DEB_VERSION)-$(TIMESTAMP_MINUTE)

# find rpm version from the spec file. latest version
# should be in the top tags, first matching 'Version: 0.1.0' and sed clears chars not in version
MUTE_RPM_VERSION := $(shell grep --only-matching --max-count 1 --line-regexp --perl-regexp "\s*Version\:\s*.+\s*" packaging/mute.spec | sed 's/[^0-9.]//g')

# command aliases
cowbuilder = env DISTRIBUTION=$(DIST) ARCH=$(ARCH) BASEPATH=/var/cache/pbuilder/base-$(DIST)-$(ARCH).cow cowbuilder


mute:
	GOOS=$(OS) GOARCH=$(GOARCH) go build -ldflags $(GOLDFLAGS) cmd/mute.go


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
	git checkout -b $(DEB_BUILD_GIT_BRANCH)
	rm -f $(MUTE_DEB_UPSTREAM_TARBAL); tar --exclude-backups --exclude-vcs -zcf $(MUTE_DEB_UPSTREAM_TARBAL) .
	cp -r packaging/debian debian; git add debian; git commit -m 'add debian dir for packaging v$(MUTE_DEB_VERSION)'
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
	tar --create --gzip --exclude-vcs --exclude=docs/man/*.rst --file $(PKG_TGZ_PATH) mute README.rst LICENSE docs/man/mute.1

# override prefix so .rpm package installs binaries to /usr/bin instead of /usr/local/bin
pkg-rpm: export prefix = /usr
# requires golang compiler > 1.13, and rpmdevtools package
pkg-rpm:
	(go version | grep -q go1.1[3-9]) || (echo "please install Go lang tools > 1.13. aborting!" && /bin/false)
	tar --exclude-vcs -zcf $(RPM_DEV_TREE)/SOURCES/mute-$(MUTE_RPM_VERSION).tar.gz .
	cp packaging/mute.spec $(RPM_DEV_TREE)/SPECS/mute-$(MUTE_RPM_VERSION).spec
	rpmbuild -bs $(RPM_DEV_TREE)/SPECS/mute-$(MUTE_RPM_VERSION).spec
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

# required: python docutils
docs:
	rst2man.py --input-encoding=utf8 --output-encoding=utf8 --strict docs/man/mute.rst docs/man/mute.1

.DEFAULT_GOAL := build
.PHONY: test build test-build install pkg-deb pkg-clean pkg-deb-setup pkg-tgz docs
