#
# spec file for package amazon-ecs-init
#
# Copyright (c) 2021 SUSE LLC
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.

# Please submit bugfixes or comments via https://bugs.opensuse.org/
#


%define short_name amazon-ecs
Name:           amazon-ecs-init
Version:        1.52.1
Release:        0
Summary:        Amazon EC2 Container Service Initialization
License:        Apache-2.0
Group:          System Environment/Base
URL:            https://github.com/aws/amazon-ecs-init
Source0:        %{name}-%{version}-1.tar.gz
Source1:        %{short_name}.service
# Patch local to openSUSE Build service to get reproducible builds
#Patch0:         reproducible.patch
# Patch local to openSUSE Build service until we sort out the cert handling
# for server validation.
#Patch1:         use-agent-container-built-in-certs.patch
BuildRequires:  go  >= 1.7
BuildRequires:  pkgconfig(systemd)
# We cannot handle cross module dependencies properly, i.e. one module can
# only depend on one other module, instead of having a one to many
# dependency construct. While docker is a hard requirement this cannot be
# expressed here and we use Recommends. AS we want to have openSUSE and SLE
# behave in the same way openSUSE has to suffer the same "brokenness"
Recommends:     docker >= 1.6.0
Requires:       systemd
Provides:       bundled(golang(github.com/Azure/go-ansiterm))
Provides:       bundled(golang(github.com/Azure/go-ansiterm/winterm))
Provides:       bundled(golang(github.com/Microsoft/go-winio))
Provides:       bundled(golang(github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml))
Provides:       bundled(golang(github.com/Nvveen/Gotty))
Provides:       bundled(golang(github.com/Sirupsen/logrus))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/awserr))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/awsutil))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/client))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/client/metadata))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/corehandlers))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/credentials))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/credentials/endpointcreds))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/credentials/stscreds))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/defaults))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/ec2metadata))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/endpoints))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/request))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/session))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/aws/signer/v4))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/internal/sdkio))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/internal/sdkrand))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/internal/shareddefaults))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/private/protocol))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/private/protocol/query))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/private/protocol/query/queryutil))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/private/protocol/rest))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/private/protocol/restxml))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/private/protocol/xml/xmlutil))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/service/s3))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/service/s3/s3iface))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/service/s3/s3manager))
Provides:       bundled(golang(github.com/aws/aws-sdk-go/service/sts))
Provides:       bundled(golang(github.com/cihub/seelog))
Provides:       bundled(golang(github.com/cihub/seelog/archive))
Provides:       bundled(golang(github.com/cihub/seelog/archive/gzip))
Provides:       bundled(golang(github.com/cihub/seelog/archive/tar))
Provides:       bundled(golang(github.com/cihub/seelog/archive/zip))
Provides:       bundled(golang(github.com/coreos/go-systemd/activation))
Provides:       bundled(golang(github.com/davecgh/go-spew/spew))
Provides:       bundled(golang(github.com/docker/docker/api/types))
Provides:       bundled(golang(github.com/docker/docker/api/types/blkiodev))
Provides:       bundled(golang(github.com/docker/docker/api/types/container))
Provides:       bundled(golang(github.com/docker/docker/api/types/filters))
Provides:       bundled(golang(github.com/docker/docker/api/types/mount))
Provides:       bundled(golang(github.com/docker/docker/api/types/network))
Provides:       bundled(golang(github.com/docker/docker/api/types/registry))
Provides:       bundled(golang(github.com/docker/docker/api/types/strslice))
Provides:       bundled(golang(github.com/docker/docker/api/types/swarm))
Provides:       bundled(golang(github.com/docker/docker/api/types/versions))
Provides:       bundled(golang(github.com/docker/docker/opts))
Provides:       bundled(golang(github.com/docker/docker/pkg/archive))
Provides:       bundled(golang(github.com/docker/docker/pkg/fileutils))
Provides:       bundled(golang(github.com/docker/docker/pkg/homedir))
Provides:       bundled(golang(github.com/docker/docker/pkg/idtools))
Provides:       bundled(golang(github.com/docker/docker/pkg/ioutils))
Provides:       bundled(golang(github.com/docker/docker/pkg/jsonlog))
Provides:       bundled(golang(github.com/docker/docker/pkg/jsonmessage))
Provides:       bundled(golang(github.com/docker/docker/pkg/longpath))
Provides:       bundled(golang(github.com/docker/docker/pkg/mount))
Provides:       bundled(golang(github.com/docker/docker/pkg/pools))
Provides:       bundled(golang(github.com/docker/docker/pkg/promise))
Provides:       bundled(golang(github.com/docker/docker/pkg/stdcopy))
Provides:       bundled(golang(github.com/docker/docker/pkg/system))
Provides:       bundled(golang(github.com/docker/docker/pkg/term))
Provides:       bundled(golang(github.com/docker/docker/pkg/term/windows))
Provides:       bundled(golang(github.com/docker/go-connections/nat))
Provides:       bundled(golang(github.com/docker/go-connections/sockets))
Provides:       bundled(golang(github.com/docker/go-plugins-helpers/sdk))
Provides:       bundled(golang(github.com/docker/go-plugins-helpers/volume))
Provides:       bundled(golang(github.com/docker/go-units))
Provides:       bundled(golang(github.com/fsouza/go-dockerclient))
Provides:       bundled(golang(github.com/go-ini/ini))
Provides:       bundled(golang(github.com/golang/mock/gomock))
Provides:       bundled(golang(github.com/jmespath/go-jmespath))
Provides:       bundled(golang(github.com/opencontainers/go-digest))
Provides:       bundled(golang(github.com/opencontainers/image-spec/specs-go))
Provides:       bundled(golang(github.com/opencontainers/image-spec/specs-go/v1))
Provides:       bundled(golang(github.com/opencontainers/runc/libcontainer/system))
Provides:       bundled(golang(github.com/opencontainers/runc/libcontainer/user))
Provides:       bundled(golang(github.com/pkg/errors))
Provides:       bundled(golang(github.com/pmezard/go-difflib/difflib))
Provides:       bundled(golang(github.com/stretchr/testify/assert))
Provides:       bundled(golang(golang.org/x/net/context))
Provides:       bundled(golang(golang.org/x/net/context/ctxhttp))
Provides:       bundled(golang(golang.org/x/net/proxy))
Provides:       bundled(golang(golang.org/x/sys/unix))
Provides:       bundled(golang(golang.org/x/sys/windows))

BuildRoot:      %{_tmppath}/%{name}-%{version}-build
ExclusiveArch:  x86_64 aarch64

%description
The Amazon EC2 Container Service initialization will start the ECS agent.
The ECS agent runs in a container and is needed to support integration
between the aws-cli ecs command line tool and an instance running in
Amazon EC2.

%prep
%setup -q -n %{name}-%{version}-1
#%patch0 -p1
#%patch1

%build
export GO111MODULE="auto"
./scripts/gobuild.sh suse
gzip -c scripts/amazon-ecs-init.1 > scripts/amazon-ecs-init.1.gz

%install
install -d -m 755 %{buildroot}/%{_mandir}/man1
install -d -m 755 %{buildroot}/%{_sbindir}
install -d -m 755 %{buildroot}/%{_sysconfdir}/ecs
install -m 644 scripts/amazon-ecs-init.1.gz %{buildroot}/%{_mandir}/man1
install -m 755 amazon-ecs-init %{buildroot}/%{_sbindir}

mkdir -p %{buildroot}/%{_unitdir}
install -m 755 %SOURCE1 %{buildroot}/%{_unitdir}

touch %{buildroot}/%{_sysconfdir}/ecs/ecs.config
touch %{buildroot}/%{_sysconfdir}/ecs/ecs.config.json

mkdir -p %{buildroot}/%{_localstatedir}/cache/ecs
touch %{buildroot}/%{_localstatedir}/cache/ecs/ecs-agent.tar
echo 0 > %{buildroot}/%{_localstatedir}/cache/ecs/state

%files
%defattr(-,root,root,-)
%dir %{_sysconfdir}/ecs
%dir %{_localstatedir}/cache/ecs
%doc CONTRIBUTING.md LICENSE NOTICE README.md
%config(noreplace) %{_sysconfdir}/ecs/ecs.config
%config(noreplace) %{_sysconfdir}/ecs/ecs.config.json
%{_mandir}/man*/*
%{_sbindir}/*
%{_unitdir}/%{short_name}.service
%{_localstatedir}/cache/ecs/ecs-agent.tar
%{_localstatedir}/cache/ecs/state

%pre
%service_add_pre %{short_name}.service

%preun
%service_del_preun %{short_name}.service

%post
%service_add_post %{short_name}.service

%postun
%service_del_postun %{short_name}.service

%changelog
