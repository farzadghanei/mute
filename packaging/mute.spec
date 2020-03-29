Name: mute
Version: 0.1.0
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

%prep
echo 'no prep needed!'


%build
%make_build


%install
rm -rf $RPM_BUILD_ROOT
%make_install
mkdir -p $RPM_BUILD_ROOT/usr/share/man/man1
cp -a docs/man/mute.1 $RPM_BUILD_ROOT/usr/share/man/man1/%{name}.1


%files
%license LICENSE
%{_bindir}/%{name}
%{_mandir}/man1/%{name}*


%changelog
* Sun Jan 05 2020 Farzad Ghanei <farzad.ghanei@tutanota.com> 0.1.0-1
- Handle signals (Closes: #4)
- Use environment variables to configure current run
- Use TOML format configuration file
