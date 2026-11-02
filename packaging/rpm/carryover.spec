Name:           carryover
Version:        1.0.0
Release:        1%{?dist}
Summary:        Linux OS Migration Application

License:        MIT
URL:            https://github.com/carryover/carryover
Source0:        carryover-%{version}.tar.gz

BuildRequires:  golang >= 1.22

%description
Carryover helps users switch Linux distributions without losing personal data.
Backs up personal files, configs, apps, and developer project rebuild recipes.
Includes cross-distro package mapping and restic encrypted backup.

%install
rm -rf $RPM_BUILD_ROOT
install -d $RPM_BUILD_ROOT/%{_bindir}
install -d $RPM_BUILD_ROOT/%{_datadir}/applications
install -d $RPM_BUILD_ROOT/%{_datadir}/icons/hicolor/scalable/apps
install -m 755 bin/carryover $RPM_BUILD_ROOT/%{_bindir}/carryover
install -m 644 packaging/carryover.desktop $RPM_BUILD_ROOT/%{_datadir}/applications/carryover.desktop
install -m 644 packaging/carryover.svg $RPM_BUILD_ROOT/%{_datadir}/icons/hicolor/scalable/apps/carryover.svg

%files
%{_bindir}/carryover
%{_datadir}/applications/carryover.desktop
%{_datadir}/icons/hicolor/scalable/apps/carryover.svg

%changelog
* Mon Nov 02 2026 Carryover Authors <authors@carryover.dev> - 1.0.0-1
- Initial release
