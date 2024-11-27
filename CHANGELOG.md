# Changelog

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
