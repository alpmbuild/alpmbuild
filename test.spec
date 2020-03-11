Name:    hello
Version: 2.10
Release: 0
Summary: Hello World
Epoch:   5

License: GPL
URL:     https://gnu.org

Source0: https://ftp.gnu.org/gnu/hello/%{name}-%{version}.tar.gz
#!alpmbuild ReasonFor cyanogen: This is a really cool package
Recommends: cyanogen
%if 0
Requires: invalid
%elif 1 == 3
Requires: still-invalid
%else
Requires: base pingas
%endif
BuildRequires: gcc
BuildRequires: make
BuildRequires: gettext

%install
mkdir -p %{?buildroot}/%{_bindir}
mkdir -p %{?buildroot}/%{_datadir}/info
mkdir -p %{?buildroot}/%{_datadir}/man/man1
mkdir -p %{?buildroot}/%{_datadir}/locale/en_US
touch %{?buildroot}/%{_bindir}/hello
touch %{?buildroot}/%{_datadir}/info/dir
touch %{?buildroot}/%{_datadir}/info/hello.info
touch %{?buildroot}/%{_datadir}/man/man1/hello.1
touch %{?buildroot}/%{_datadir}/locale/en_US/LC_MESSAGES/hello.mo
chown -R root:root %{?buildroot}

%files
%{_bindir}/hello
%{_datadir}/info/dir
%{_datadir}/info/hello.info
%{_datadir}/man/man1/hello.1

%package translationfiles
Summary: Translation files for %{name}

%files translationfiles
%{_datadir}/locale/*/LC_MESSAGES/hello.mo