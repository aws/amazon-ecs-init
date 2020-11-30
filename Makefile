# Copyright 2014-2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the
# "License"). You may not use this file except in compliance
#  with the License. A copy of the License is located at
#
#     http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is
# distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
# CONDITIONS OF ANY KIND, either express or implied. See the
# License for the specific language governing permissions and
# limitations under the License.
VERSION := $(shell git describe --tags | sed -e 's/v//' -e 's/-.*//')
DEB_SIGN ?= 1

.PHONY: dev generate lint static test build-mock-images sources rpm srpm govet

dev:
	./scripts/gobuild.sh dev

generate:
	PATH=$(PATH):$(shell pwd)/scripts && cd ecs-init && $(MAKE) generate

static:
	./scripts/gobuild.sh

govet:
	cd ecs-init && $(MAKE) govet

test:
	cd ecs-init && $(MAKE) test

.PHONY: analyze-cover-profile
analyze-cover-profile: ./ecs-init/coverprofile.out
	./scripts/analyze-cover-profile

.PHONY: gocyclo
gocyclo:
	cd ecs-init && $(MAKE) gocyclo

.PHONY: goimports
goimports:
	@cd ecs-init && $(MAKE) goimports

.PHONY: importcheck
importcheck:
	cd ecs-init && $(MAKE) importcheck

.PHONY: static-check
static-check:
	cd ecs-init && $(MAKE) static-check

test-in-docker:
	docker build -f scripts/dockerfiles/test.dockerfile -t "amazon/amazon-ecs-init-test:make" .
	docker run -v "$(shell pwd):/go/src/github.com/aws/amazon-ecs-init" "amazon/amazon-ecs-init-test:make"

build-mock-images:
	docker build -t "test.localhost/amazon/mock-ecs-agent" -f "scripts/dockerfiles/mock-agent.dockerfile" .
	docker build -t "test.localhost/amazon/wants-update" -f "scripts/dockerfiles/wants-update.dockerfile" .
	docker build -t "test.localhost/amazon/exit-success" -f "scripts/dockerfiles/exit-success.dockerfile" .

sources.tgz:
	./scripts/update-version.sh
	cp packaging/amazon-linux-ami/ecs-init.spec ecs-init.spec
	cp packaging/amazon-linux-ami/ecs.conf ecs.conf
	cp packaging/amazon-linux-ami/ecs.service ecs.service
	cp packaging/amazon-linux-ami/amazon-ecs-volume-plugin.conf amazon-ecs-volume-plugin.conf
	cp packaging/amazon-linux-ami/amazon-ecs-volume-plugin.service amazon-ecs-volume-plugin.service
	cp packaging/amazon-linux-ami/amazon-ecs-volume-plugin.socket amazon-ecs-volume-plugin.socket
	tar -czf ./sources.tgz ecs-init scripts

# Hook to perform preparation steps prior to the sources target.
prepare-sources::

sources: prepare-sources sources.tgz

.srpm-done: sources.tgz
	test -e SOURCES || ln -s . SOURCES
	rpmbuild --define "%_topdir $(PWD)" -bs ecs-init.spec
	find SRPMS/ -type f -exec cp {} . \;
	touch .srpm-done

srpm: .srpm-done

.rpm-done: sources.tgz
	test -e SOURCES || ln -s . SOURCES
	rpmbuild --define "%_topdir $(PWD)" -bb ecs-init.spec
	find RPMS/ -type f -exec cp {} . \;
	touch .rpm-done

rpm: .rpm-done

ubuntu-trusty:
	cp packaging/ubuntu-trusty/ecs.conf ecs.conf
	tar -czf ./amazon-ecs-init_${VERSION}.orig.tar.gz ecs-init ecs.conf scripts README.md
	mkdir -p BUILDROOT
	cp -r packaging/ubuntu-trusty/debian BUILDROOT/debian
	cp -r ecs-init BUILDROOT
	cp packaging/ubuntu-trusty/ecs.conf BUILDROOT
	cp -r scripts BUILDROOT
	cp README.md BUILDROOT
	cd BUILDROOT && debuild $(shell [ "$(DEB_SIGN)" -ne "0" ] || echo "-uc -us")

get-deps:
	cd ecs-init && go list -f '{{ join .Imports "\n" }}' tools.go | xargs -tI % go install %

clean:
	-rm -f ecs-init.spec
	-rm -f ecs.conf
	-rm -f ecs.service
	-rm -f amazon-ecs-volume-plugin.conf
	-rm -f amazon-ecs-volume-plugin.service
	-rm -f amazon-ecs-volume-plugin.socket
	-rm -f ./sources.tgz
	-rm -f ./amazon-ecs-init
	-rm -f ./ecs-agent-*.tar
	-rm -f ./ecs-init-*.src.rpm
	-rm -rf ./ecs-init-*
	-rm -rf ./BUILDROOT BUILD RPMS SRPMS SOURCES SPECS
	-rm -rf ./x86_64
	-rm -f ./amazon-ecs-init_${VERSION}*
	-rm -f .srpm-done .rpm-done
	-rm -f ./ecs-init/cover.out
	-rm -f ./ecs-init/coverprofile.out
