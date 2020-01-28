#!alpmbuild NoFileCheck
Name:    hello
Version: 2.10
Release: 0
Summary: Hello World
License: GPL
URL:     https://gnu.org
Source0: https://ftp.gnu.org/gnu/hello/%{name}-%{version}.tar.gz

%prep
tar -xvf %{name}-%{version}.tar.gz
cd %{name}-%{version}

%build
./configure --prefix=%{_prefix}
make

%install
make DESTDIR=$PREFIX install