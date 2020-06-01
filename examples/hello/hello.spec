Name:    hello
Version: 2.10
Summary: Hello World
License: GPL
URL:     https://gnu.org
Source0: https://ftp.gnu.org/gnu/hello/%{name}-%{version}.tar.gz \
         with md5 6cd0ffea3884a4e79330338dcc2987d6

%build
./configure --prefix=%{_prefix}
make

%install
make install DESTDIR=%{buildroot}