%define debug_package %{nil}
%global provider        github
%global provider_tld    com
%global project         kubernetes-incubator
%global repo            cri-o
%global Name            ocid
# https://github.com/kubernetes-incubator/cri-o
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path     %{provider_prefix}
%global commit          8ba639952a95f2e24cc98987689138b67545576c
%global shortcommit     %(c=%{commit}; echo ${c:0:7})

Name:           %{Name}
Version:        0.0.1
Release:        1.git%{shortcommit}%{?dist}
Summary:        Kubelet Container Runtime Interface (CRI) for OCI runtimes.
Group:          Applications/Text
License:        Apache 2.0
URL:            https://%{provider_prefix}
Source0:        https://%{provider_prefix}/archive/%{commit}/%{repo}-%{shortcommit}.tar.gz

BuildRequires:  golang-github-cpuguy83-go-md2man

%description
The ocid package provides an implementation of the
Kubelet Container Runtime Interface (CRI) using OCI conformant runtimes.

ocid provides following functionalities:

    Support multiple image formats including the existing Docker image format
    Support for multiple means to download images including trust & image verification
    Container image management (managing image layers, overlay filesystems, etc)
    Container process lifecycle management
    Monitoring and logging required to satisfy the CRI
    Resource isolation as required by the CRI

%prep
%setup -q -n %{repo}-%{commit}

%build
make all

%install
%make_install
%make_install install.systemd

#define license tag if not already defined
%{!?_licensedir:%global license %doc}
%files
%{_bindir}/ocid
%{_bindir}/ocic
%{_mandir}/man5/ocid.conf.5*
%{_mandir}/man8/ocid.8*
%{_sysconfdir}/ocid.conf
%dir /%{_libexecdir}/ocid
/%{_libexecdir}/ocid/conmon
/%{_libexecdir}/ocid/pause
%{_unitdir}/ocid.service
%doc README.md
%license LICENSE

%preun
%systemd_preun %{Name}

%postun
%systemd_postun_with_restart %{Name}

%changelog
* Mon Oct 31 2016 Dan Walsh <dwalsh@redhat.com> - 0.0.1
- Initial RPM release

