#!alpmbuild NoFileCheck
Name: yay
Version: 9.4.4
Release: 1
Summary: Yet another yogurt. Pacman wrapper and AUR helper written in go. In a specfile, because of course the go program is packaged by another go program.
URL: https://github.com/Jguer/yay
License: GPL
Requires: pacman>=5.2 sudo git
BuildRequires: go
Source0: https://github.com/Jguer/yay/archive/v%{version}.tar.gz

%prep
tar -xvf v%{version}.tar.gz
cd %{name}-%{version}

%build
make VERSION=%{version} DESTDIR=$PREFIX build

%install
make VERSION=%{version} DESTDIR=$PREFIX PREFIX=%{_prefix} install