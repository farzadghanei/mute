Name: mute
Version: 0.3.0
Release: 1%{?dist}
Summary: Run other programs muting the output when configured

License: MIT
URL: https://github.com/farzadghanei/mute
Source0: %{name}-%{version}.tar.gz

# will use official golang tarballs instead, until go 1.13 rpm is in most repos
# BuildRequires: golang > 1.12.5, golang-github-BurntSushi-toml > 0.3.1

%description
mute runs other programs and mutes the output under configured
conditions. A good use case is to keep cron jobs silenced and avoid receiving
emails for known conditions.


# go toolchain stores go build id in a different ELF note than GNU toolchain
# so RPM can't find the build id from the binaries after build.
# https://github.com/rpm-software-management/rpm/issues/367
%global _missing_build_ids_terminate_build 0
%define debug_package %{nil}

%prep
%setup -c -q

%build
%make_build


%install
rm -rf $RPM_BUILD_ROOT
%make_install
mkdir -p $RPM_BUILD_ROOT/usr/share/man/man1
cp -a docs/man/mute.1 $RPM_BUILD_ROOT/usr/share/man/man1/%{name}.1


%clean
rm -rf $RPM_BUILD_ROOT


%files
%license LICENSE
%doc README.rst
%{_bindir}/%{name}
%{_mandir}/man1/%{name}*


%changelog

* Sat May 24 2025 Farzad Ghanei <644113+farzadghanei@users.noreply.github.com> 0.3.0-1
- Update dependencies (toml from 0.3.1 to 1.5.0, go from 1.13 to 1.24)

* Wed Nov 04 2020 Farzad Ghanei <644113+farzadghanei@users.noreply.github.com> 0.2.0-1
- Restructure project layout
- Restructure Exec, reduce chances of memory allocation failures (Closes: #19)
- Exec supports pre allocating buffers to prevent extra allocation later
- Rename constants ENV_* to CamelCase as a more conventional format

* Sun Aug 16 2020 Farzad Ghanei <644113+farzadghanei@users.noreply.github.com> 0.1.1-1
- Fix missing new line at the end of the help message (Closes: #15)
- Fix formatting issues in README
- Redo Debian packaging, support Debian buster, use gbp (Closes: #12)
- Add RPM packaging

* Sun Jan 05 2020 Farzad Ghanei <644113+farzadghanei@users.noreply.github.com> 0.1.0-1
- Handle signals (Closes: #4)
- Use environment variables to configure current run
- Use TOML format configuration file
