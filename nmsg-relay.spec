%undefine _missing_build_ids_terminate_build
%global debug_package   %{nil}
%global provider        github
%global provider_tld    com
%global project         farsightsec
%global repo            nmsg-relay
# https://github.com/farsightsec/nmsg-relay
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path     %{provider_prefix}
%global commit          16fcd3b05a8e2cf8d238c2699b0fc9fb588f4107
%global shortcommit     %(c=%{commit}; echo ${c:0:7})

Name:           nmsg-relay
Version:        0.2.0
Release:	1%{?dist}
Summary:        SIE uploader for NMSG data
License:        MPLv2.0
URL:            https://%{provider_prefix}
Source0:        https://%{provider_prefix}/archive/%{commit}/%{repo}-%{shortcommit}.tar.gz

BuildRequires:  %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}
BuildRequires: golang-github-farsightsec-go-nmsg-devel
#sielink end with devel
BuildRequires: golang-github-farsightsec-sielink
BuildRequires: golang-github-golang-protobuf-devel	
#full long name
BuildRequires: go-config-devel


%if %{rhel} == 9 
#yaml.v2-devel
BuildRequires: golang-gopkg-yaml-devel-v2

%else

BuildRequires: golang-gopkg-yaml-devel-v2 

%endif

%description

%{summary}

A lightweight client which reads NMSG input from a datagram socket and submits it to the SIE.

%prep
%setup -q -n %{repo}-%{commit}

%build
mkdir -p /builddir/go/src/github.com/farsightsec
ln -s $PWD /builddir/go/src/github.com/farsightsec/nmsg-relay

%{!?_licensedir:%global license %doc}

export GO111MODULE=off 
export GOPATH=/usr/share/gocode:/builddir/go 
go build 


%install
install -d -p %{buildroot}%{_bindir}
#install %{repo}-%{version} %{buildroot}/%{_bindir}/nmsg-relay
install ./nmsg-relay-%{commit} %{buildroot}/%{_bindir}/nmsg-relay
install -d -p %{buildroot}%{_mandir}/man1
install nmsg-relay.1 %{buildroot}%{_mandir}/man1/

%files
%license LICENSE
%doc README.md COPYRIGHT
%{_bindir}/nmsg-relay
%_mandir/man1/*

%changelog
