# Changelog

## [0.13.0](https://github.com/runfinch/finch-daemon/compare/v0.12.0...v0.13.0) (2025-02-18)


### Build System or External Dependencies

* **deps:** Bump github.com/onsi/ginkgo/v2 from 2.22.1 to 2.22.2 ([#138](https://github.com/runfinch/finch-daemon/issues/138)) ([776946b](https://github.com/runfinch/finch-daemon/commit/776946b2de620d679980083596486446d41a8b73))
* **deps:** Bump github.com/runfinch/common-tests from 0.8.0 to 0.9.1 ([#140](https://github.com/runfinch/finch-daemon/issues/140)) ([96b109e](https://github.com/runfinch/finch-daemon/commit/96b109e29c592a36dbe39f81134669f6f9525249))
* **deps:** Bump github.com/spf13/afero from 1.11.0 to 1.12.0 ([#150](https://github.com/runfinch/finch-daemon/issues/150)) ([db169fb](https://github.com/runfinch/finch-daemon/commit/db169fb2fc86abf10b6af856b94238bd7729b573))
* **deps:** Bump golang.org/x/net from 0.33.0 to 0.34.0 ([#141](https://github.com/runfinch/finch-daemon/issues/141)) ([5b1fe1c](https://github.com/runfinch/finch-daemon/commit/5b1fe1ca1f7a00356bd24f5f85367f24f242e6b7))
* **deps:** Bump google.golang.org/protobuf from 1.36.1 to 1.36.4 ([#151](https://github.com/runfinch/finch-daemon/issues/151)) ([f7c7b30](https://github.com/runfinch/finch-daemon/commit/f7c7b3070d141c7d90405a230d4a48062fde56ac))


### Features

* Add container kill API ([#146](https://github.com/runfinch/finch-daemon/issues/146)) ([8a40617](https://github.com/runfinch/finch-daemon/commit/8a4061717384385c51910f9e8522a0268e6238ed))
* Update container inspect with size option ([#157](https://github.com/runfinch/finch-daemon/issues/157)) ([b5df6ef](https://github.com/runfinch/finch-daemon/commit/b5df6ef819803699cc11fccda7c0598b5672af3e))

### Others

* Update containerd and nerdctl to v2 ([#148](https://github.com/runfinch/finch-daemon/pull/148)) ([d3db35e](https://github.com/runfinch/finch-daemon/commit/d3db35eae8b825732fc1b6c08960b63779f5e92a))

## [0.12.0](https://github.com/runfinch/finch-daemon/compare/v0.11.0...v0.12.0) (2024-12-27)


### Build System or External Dependencies

* **deps:** Bump github.com/containerd/cgroups/v3 from 3.0.3 to 3.0.5 ([#130](https://github.com/runfinch/finch-daemon/issues/130)) ([bc69841](https://github.com/runfinch/finch-daemon/commit/bc6984128b4c47bb933f6aba46b7802d9bf9e70d))
* **deps:** Bump github.com/containerd/go-cni from 1.1.10 to 1.1.11 ([#134](https://github.com/runfinch/finch-daemon/issues/134)) ([82f9629](https://github.com/runfinch/finch-daemon/commit/82f9629b3088450c7bf6ef178344993a44aea785))
* **deps:** Bump github.com/containerd/typeurl/v2 from 2.2.0 to 2.2.3 ([#133](https://github.com/runfinch/finch-daemon/issues/133)) ([d9704ff](https://github.com/runfinch/finch-daemon/commit/d9704ff415483774bdffe1b474b4d25315346abd))
* **deps:** Bump github.com/docker/cli from 26.0.0+incompatible to 27.4.1+incompatible ([#125](https://github.com/runfinch/finch-daemon/issues/125)) ([bff4a1f](https://github.com/runfinch/finch-daemon/commit/bff4a1fe3eae771e491583e99dad3d14ee11d97c))
* **deps:** Bump github.com/docker/docker from 26.1.5+incompatible to 27.4.1+incompatible ([#127](https://github.com/runfinch/finch-daemon/issues/127)) ([f77066d](https://github.com/runfinch/finch-daemon/commit/f77066dd388b1290dc4c31c3f8e28301c0b96f8d))
* **deps:** Bump github.com/moby/moby from 26.0.0+incompatible to 27.4.1+incompatible ([#132](https://github.com/runfinch/finch-daemon/issues/132)) ([da5f7cd](https://github.com/runfinch/finch-daemon/commit/da5f7cd3339d3164fd1c51fe8b55bd42d4c92757))
* **deps:** Bump github.com/onsi/ginkgo/v2 from 2.20.2 to 2.22.1 ([#129](https://github.com/runfinch/finch-daemon/issues/129)) ([606c16b](https://github.com/runfinch/finch-daemon/commit/606c16b0ccbf018aa65cdc4b4c3cbc3e0071061c))
* **deps:** Bump github.com/onsi/gomega from 1.36.1 to 1.36.2 ([#136](https://github.com/runfinch/finch-daemon/issues/136)) ([72facd2](https://github.com/runfinch/finch-daemon/commit/72facd29138b19bae5302fc06efd281a26c28cdd))
* **deps:** Bump github.com/vishvananda/netns from 0.0.4 to 0.0.5 ([#112](https://github.com/runfinch/finch-daemon/issues/112)) ([40d3991](https://github.com/runfinch/finch-daemon/commit/40d3991864516be6816abbb788118aaaf6e29802))
* **deps:** Bump golang.org/x/crypto from 0.29.0 to 0.31.0 ([#120](https://github.com/runfinch/finch-daemon/issues/120)) ([5c99f3e](https://github.com/runfinch/finch-daemon/commit/5c99f3e7ddfacdf5793c7eac32c6d7e1aa54d780))
* **deps:** Bump google.golang.org/protobuf from 1.35.1 to 1.35.2 ([#116](https://github.com/runfinch/finch-daemon/issues/116)) ([682d4cd](https://github.com/runfinch/finch-daemon/commit/682d4cddf09c19a254e5341d9c7842d4da3e2d3c))
* **deps:** Bump google.golang.org/protobuf from 1.35.2 to 1.36.0 ([#128](https://github.com/runfinch/finch-daemon/issues/128)) ([b46d499](https://github.com/runfinch/finch-daemon/commit/b46d49935a6001c63f884c1d184346ea7445a8b8))


### Features

* add distribution API (with bug fix) ([#121](https://github.com/runfinch/finch-daemon/issues/121)) ([da0dab7](https://github.com/runfinch/finch-daemon/commit/da0dab73d84aac1d7e7ac432774480d49ccf495e))
* add more options to container create API ([#122](https://github.com/runfinch/finch-daemon/pull/122)) ([9fda9cd](https://github.com/runfinch/finch-daemon/commit/9fda9cd9d4ae6355588410871b71ef5217082bbb))


### Bug Fixes

* Update go mod to fix CVE-2024-45338 ([#124](https://github.com/runfinch/finch-daemon/issues/124)) ([19a3980](https://github.com/runfinch/finch-daemon/commit/19a3980b0bd897fb814143eeb8255837f42c1014))

## [0.11.0](https://github.com/runfinch/finch-daemon/compare/v0.10.0...v0.11.0) (2024-11-27)


### Build System or External Dependencies

* **deps:** Bump github.com/containernetworking/cni from 1.2.2 to 1.2.3 ([#87](https://github.com/runfinch/finch-daemon/issues/87)) ([46df1b6](https://github.com/runfinch/finch-daemon/commit/46df1b631d6d9cf77d1a34b1162c9ac0226e5ff6))
* **deps:** Bump github.com/coreos/go-iptables from 0.7.0 to 0.8.0 ([#106](https://github.com/runfinch/finch-daemon/issues/106)) ([9905569](https://github.com/runfinch/finch-daemon/commit/990556941eee136457e190de217e9e64249b54d1))
* **deps:** Bump github.com/stretchr/testify from 1.9.0 to 1.10.0 ([#107](https://github.com/runfinch/finch-daemon/issues/107)) ([e5b9878](https://github.com/runfinch/finch-daemon/commit/e5b987880954c14b90a2f984a586b2b02eeec44c))
* **deps:** Bump golang.org/x/net from 0.29.0 to 0.31.0 ([#93](https://github.com/runfinch/finch-daemon/issues/93)) ([dcffe39](https://github.com/runfinch/finch-daemon/commit/dcffe399140761198e2fec7de08005a7c56c5c3f))
* **deps:** Bump github.com/containerd/containerd from 1.7.22 to 1.7.24 ([#102](https://github.com/runfinch/finch-daemon/issues/102)) ([0d6cd12](https://github.com/runfinch/finch-daemon/commit/0d6cd122af858ed4431ebf37a673ad933054c833))
* **deps:** Bump github.com/containerd/errdefs from 0.1.0 to 1.0.0 ([#102](https://github.com/runfinch/finch-daemon/issues/102)) ([0d6cd12](https://github.com/runfinch/finch-daemon/commit/0d6cd122af858ed4431ebf37a673ad933054c833))


### Features

* Implementation of enable_icc option ([#69](https://github.com/runfinch/finch-daemon/issues/69)) ([5fd2e3e](https://github.com/runfinch/finch-daemon/commit/5fd2e3ee7cf1f17f59c58028fd931bc9a9f51b38))


### Bug Fixes

* Make DOCKER_CONFIG available to buildctl ([#94](https://github.com/runfinch/finch-daemon/issues/94)) ([f5b426d](https://github.com/runfinch/finch-daemon/commit/f5b426d058c8700e4a4111143db131b4582287d8))
* Pidfile handling and socket docs ([#101](https://github.com/runfinch/finch-daemon/issues/101)) ([5c2e99f](https://github.com/runfinch/finch-daemon/commit/5c2e99f22388d184b2f7916432cac1173622143c))
* return an error if custom bridge name is not set successfully ([#100](https://github.com/runfinch/finch-daemon/issues/100)) ([0469999](https://github.com/runfinch/finch-daemon/commit/0469999c87b8659b149617cc99ab919e1a09b752))


## [0.10.0](https://github.com/runfinch/finch-daemon/compare/v0.9.0...v0.10.0) (2024-10-31)


### Build System or External Dependencies

* **deps:** bump github.com/containerd/nerdctl from 1.7.5 to 1.7.7 ([#66](https://github.com/runfinch/finch-daemon/issues/66)) ([80fdae9](https://github.com/runfinch/finch-daemon/commit/80fdae9e466a2df51f61f6f7ab22effe21f5913f))
* **deps:** bump github.com/runfinch/common-tests from 0.7.21 to 0.8.0 ([#64](https://github.com/runfinch/finch-daemon/issues/64)) ([df9f0ca](https://github.com/runfinch/finch-daemon/commit/df9f0cad2f1cc842a6c3033dc2d635008a2690df))


### Features

* Add Support for Extra Hosts ([#85](https://github.com/runfinch/finch-daemon/issues/85)) ([5722300](https://github.com/runfinch/finch-daemon/commit/5722300912f8a4cdcc4aa22bae6524ef79a9b7d1))
* Add support for nerdctl config and default variables ([#73](https://github.com/runfinch/finch-daemon/issues/73)) ([284c73f](https://github.com/runfinch/finch-daemon/commit/284c73ffc02ac5bd1712b92e06675474cb206c19))
* Add support for pidfile ([#90](https://github.com/runfinch/finch-daemon/issues/90)) ([55eacb5](https://github.com/runfinch/finch-daemon/commit/55eacb5f8ed302bf8aa2138a9b47b2c01970e28b))
* Add support for socket Activation ([#89](https://github.com/runfinch/finch-daemon/issues/89)) ([d185ad3](https://github.com/runfinch/finch-daemon/commit/d185ad3b2fc057fb7655ee0168d4ffea679df432))


### Bug Fixes

* Add static binaries to release ([#63](https://github.com/runfinch/finch-daemon/issues/63)) ([57a0c44](https://github.com/runfinch/finch-daemon/commit/57a0c44d56bbf0addbf5b8c78a2baebac61141ab))
* bump minor for feat ([65fbf5a](https://github.com/runfinch/finch-daemon/commit/65fbf5afaeb175d5660ff13acc639ec3d72ac273))
* Remove bump-patch-for-minor-pre-major ([#83](https://github.com/runfinch/finch-daemon/issues/83)) ([65fbf5a](https://github.com/runfinch/finch-daemon/commit/65fbf5afaeb175d5660ff13acc639ec3d72ac273))

## 0.9.0  Finch-Daemon Init

This is the first release of the Finch Daemon.
The Finch Daemon project is an open source container runtime engine that enables users to integrate software that uses Docker's RESTful APIs as a programmatic dependency.
