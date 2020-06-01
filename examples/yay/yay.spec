Name: yay
Version: 9.4.4
Summary: Yet another yogurt. Pacman wrapper and AUR helper written in go.
URL: https://github.com/Jguer/yay
License: GPL
Requires: pacman>=5.2 sudo git
BuildRequires: go
Source0: https://github.com/Jguer/yay/archive/v%{version}.tar.gz \
         with sha1 da952b34a9bf833d71a7403c394b758587c1504e

%build
export CGO_CPPFLAGS="${CPPFLAGS}"
export CGO_CFLAGS="${CFLAGS}"
export CGO_CXXFLAGS="${CXXFLAGS}"
export CGO_LDFLAGS="${LDFLAGS}"
export GOFLAGS="-buildmode=pie -trimpath -mod=readonly -modcacherw"
make VERSION=%{version} DESTDIR=$PREFIX build

%install
make VERSION=%{version} DESTDIR=$PREFIX PREFIX=%{_prefix} install