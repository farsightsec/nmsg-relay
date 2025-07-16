# Define backup go macros
%if %{rhel} == 8
%global gopkg %package -n %{goname}-devel \
Summary:	%{summary} \
BuildArch:  noarch \
%description -n %{goname}-devel \
%{common_description}
%global goprep(A) %setup -q
%global generate_buildrequires echo "Need more specific macro on rhel8"
%global gopkginstall for file in $(find . -iname "*.go" \! -iname "*_test.go" \! -iname "main.go" ) ; do \
    echo "%%dir %%{gopath}/src/%%{goipath}/$(dirname $file)" >> devel.file-list ;\
    install -d -p %{buildroot}/%{gopath}/src/%{goipath}/$(dirname $file) ;\
    cp -pav $file %{buildroot}/%{gopath}/src/%{goipath}/$file ;\
    echo "%%{gopath}/src/%%{goipath}/$file" >> devel.file-list ;\
done ;\
sort -u -o devel.file-list devel.file-list
%global gopkgfiles %files -n %{goname}-devel -f devel.file-list
%global gocheck echo "skipping gocheck on rhel8"
# Specific BuildRequires macro
%global go_generate_buildrequires BuildRequires: %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang} \
BuildRequires: golang-github-dnstap-devel \
BuildRequires: golang-github-farsightsec-config-devel \
BuildRequires: golang-github-farsightsec-framestream-devel \
BuildRequires: golang-github-farsightsec-nmsg-sie-devel \
BuildRequires: golang-github-farsightsec-nmsg-devel \
BuildRequires: golang-github-farsightsec-sielink-devel \
BuildRequires: golang-github-miekg-dns-devel \
BuildRequires: golang-github-protobuf-devel \
BuildRequires: golang-google-protobuf-devel \
BuildRequires: golang-gopkg-yaml-2-devel \
BuildRequires: golang-x-net-devel \
BuildRequires: golang-x-sys-devel
%endif

%global debug_package %{nil}
%define _build_id_links none
# https://github.com/farsightsec/nmsg-relay
%global goipath         github.com/farsightsec/nmsg-relay
%global common_description %{expand:
%{summary}

A lightweight client which reads NMSG input from a datagram socket and submits it to the SIE.}

Name:           nmsg-relay
Version:        0.2.0
Release:        1%{?dist}
Summary:        SIE uploader for NMSG data
%gometa
License:        MPLv2.0
URL:            %{gourl}
Source0:        %{gosource}

%description
%{common_description}

%gopkg

%generate_buildrequires
%goprep -A
%autopatch -p1
mkdir -p /builddir/go/src/github.com/farsightsec
ln -s $PWD /builddir/go/src/github.com/farsightsec/nmsg-relay
%go_generate_buildrequires

%prep
%goprep -A
%autopatch -p1

%build
mkdir -p /builddir/go/src/github.com/farsightsec
ln -s $PWD /builddir/go/src/github.com/farsightsec/nmsg-relay

%{!?_licensedir:%global license %doc}

# We don't want to download new modules, but we want ones that BuildRequires packages provide
export GO111MODULE=off
export GOPATH=/usr/share/gocode:/builddir/go
go build

%install
install -d -p %{buildroot}%{_bindir}
install %{name}-%{version} %{buildroot}/%{_bindir}/nmsg-relay
install -d -p %{buildroot}%{_mandir}/man1
install nmsg-relay.1 %{buildroot}%{_mandir}/man1/
%gopkginstall


%if %{with check}
%check
%gocheck
%endif

#define license tag if not already defined
%{!?_licensedir:%global license %doc}

%files
%license LICENSE
%doc README.md COPYRIGHT
%{_bindir}/nmsg-relay
%_mandir/man1/*
%gopkgfiles

%changelog
