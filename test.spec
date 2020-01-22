Name:    hello
Version: 2.10
Release: 0
Summary: Hello World

License: GPL
URL:     https://gnu.org

Source0: https://ftp.gnu.org/gnu/hello/%{name}-%{version}.tar.gz

BuildRequires: gcc make gettext

%prep
%autosetup

%build
%configure
%make_build

%install
%make_install

%files
%{_bindir}/hello
%{_datadir}/info/hello.info.gz
%{_datadir}/man/man1/hello.1.gz
%{_datadir}/locale/*/LC_MESSAGES/hello.mo