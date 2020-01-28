Name:    hello
Version: 2.10
Release: 0
Summary: Hello World

License: GPL
URL:     https://gnu.org

Source0: https://ftp.gnu.org/gnu/hello/%{name}-%{version}.tar.gz

%if 1 == 2
Requires: invalid
%elif 1 == 3
Requires: still-invalid
%else
Requires: base
%endif
BuildRequires: gcc
BuildRequires: make
BuildRequires: gettext

%prep
tar -xvf %{name}-%{version}.tar.gz
cd %{name}-%{version}

%build
./configure --prefix=%{_prefix}
make

%install
make DESTDIR=$PREFIX install

%files
%{_bindir}/hello
%{_datadir}/info/dir
%{_datadir}/info/hello.info
%{_datadir}/man/man1/hello.1

%package translationfiles
Summary: Translation files for %{name}

%files translationfiles
%{_datadir}/locale/*/LC_MESSAGES/hello.mo