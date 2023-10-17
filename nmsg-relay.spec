%global debug_package %{nil}

# https://github.com/farsightsec/nmsg-relay
%global goipath         github.com/farsightsec/nmsg-relay
Version:                0.3.0

%gometa

%global common_description %{expand:
A lightweight client which reads NMSG input from a datagram socket and submits it to the SIE.}

%global golicenses      LICENSE
%global godocs          README.md

Name:           nmsg-relay
Release:        %autorelease
Summary:        NMSG relay to SIE

License:        MPLv2.0
URL:            %{gourl}
Source0:        %{gosource}

%description
%{common_description}

%gopkg

%prep
%goprep

%generate_buildrequires
%go_generate_buildrequires

%build
for cmd in . ; do
  %gobuild -o %{gobuilddir}/bin/$(basename $cmd) %{goipath}/$cmd
done

%install
%gopkginstall
install -m 0755 -vd                     %{buildroot}%{_bindir}
install -m 0755 -vp %{gobuilddir}/bin/* %{buildroot}%{_bindir}/

%if %{with check}
%check
%gocheck
%endif

%files
%doc     
%{_bindir}/*

%gopkgfiles

%changelog
%autochangelog
