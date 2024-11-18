# Changelog

## [0.11.0](https://github.com/Shubhranshu153/finch-daemon/compare/v0.10.0...v0.11.0) (2024-11-18)


### Build System or External Dependencies

* **deps:** bump github.com/containerd/go-cni from 1.1.9 to 1.1.10 ([#53](https://github.com/Shubhranshu153/finch-daemon/issues/53)) ([31583b0](https://github.com/Shubhranshu153/finch-daemon/commit/31583b0bd25dfdcf5c53ae78882b9df3ac36cc11))
* **deps:** bump github.com/containerd/nerdctl from 1.7.5 to 1.7.7 ([#66](https://github.com/Shubhranshu153/finch-daemon/issues/66)) ([80fdae9](https://github.com/Shubhranshu153/finch-daemon/commit/80fdae9e466a2df51f61f6f7ab22effe21f5913f))
* **deps:** bump github.com/onsi/ginkgo/v2 from 2.17.1 to 2.20.2 ([#19](https://github.com/Shubhranshu153/finch-daemon/issues/19)) ([e282c25](https://github.com/Shubhranshu153/finch-daemon/commit/e282c253bfdd2bad7e97866e75598291892fb7fa))
* **deps:** bump github.com/onsi/gomega from 1.32.0 to 1.34.2 ([#18](https://github.com/Shubhranshu153/finch-daemon/issues/18)) ([ea72df3](https://github.com/Shubhranshu153/finch-daemon/commit/ea72df3f479e10ef0de0357a31a1686d626f5041))
* **deps:** bump github.com/runfinch/common-tests from 0.7.21 to 0.8.0 ([#64](https://github.com/Shubhranshu153/finch-daemon/issues/64)) ([df9f0ca](https://github.com/Shubhranshu153/finch-daemon/commit/df9f0cad2f1cc842a6c3033dc2d635008a2690df))
* **deps:** bump github.com/spf13/cobra from 1.8.0 to 1.8.1 ([#49](https://github.com/Shubhranshu153/finch-daemon/issues/49)) ([3eff666](https://github.com/Shubhranshu153/finch-daemon/commit/3eff666f81e4ea655b9d70e5fa7e8043283ec959))
* **deps:** bump github.com/vishvananda/netlink from 1.2.1-beta.2 to 1.3.0 ([#50](https://github.com/Shubhranshu153/finch-daemon/issues/50)) ([e3cffc7](https://github.com/Shubhranshu153/finch-daemon/commit/e3cffc77ac28451c15d5c6a04ab63fd89c34fe4b))


### Features

* add container create options ([#27](https://github.com/Shubhranshu153/finch-daemon/issues/27)) ([504dcaf](https://github.com/Shubhranshu153/finch-daemon/commit/504dcaf9eff1316c9dd40db82a4ecce9b3e1796d))
* add distribution API ([#92](https://github.com/Shubhranshu153/finch-daemon/issues/92)) ([0e413d7](https://github.com/Shubhranshu153/finch-daemon/commit/0e413d7a3833f2b392921bf7131e80bf6b969fa0))
* Add Support for Extra Hosts ([#85](https://github.com/Shubhranshu153/finch-daemon/issues/85)) ([5722300](https://github.com/Shubhranshu153/finch-daemon/commit/5722300912f8a4cdcc4aa22bae6524ef79a9b7d1))
* Add support for nerdctl config and default variables ([#73](https://github.com/Shubhranshu153/finch-daemon/issues/73)) ([284c73f](https://github.com/Shubhranshu153/finch-daemon/commit/284c73ffc02ac5bd1712b92e06675474cb206c19))
* Add support for pidfile ([#90](https://github.com/Shubhranshu153/finch-daemon/issues/90)) ([55eacb5](https://github.com/Shubhranshu153/finch-daemon/commit/55eacb5f8ed302bf8aa2138a9b47b2c01970e28b))
* Add support for socket Activation ([#89](https://github.com/Shubhranshu153/finch-daemon/issues/89)) ([d185ad3](https://github.com/Shubhranshu153/finch-daemon/commit/d185ad3b2fc057fb7655ee0168d4ffea679df432))
* allow custom socket path ([#7](https://github.com/Shubhranshu153/finch-daemon/issues/7)) ([4c17545](https://github.com/Shubhranshu153/finch-daemon/commit/4c1754576d5beb3bd6b12e36893a588b2bb95825))
* implement container restart API ([#23](https://github.com/Shubhranshu153/finch-daemon/issues/23)) ([5d9b1e0](https://github.com/Shubhranshu153/finch-daemon/commit/5d9b1e0f4e1565fd374b0f0941f373a094dc749c))
* Implementation of enable_icc option ([#69](https://github.com/Shubhranshu153/finch-daemon/issues/69)) ([5fd2e3e](https://github.com/Shubhranshu153/finch-daemon/commit/5fd2e3ee7cf1f17f59c58028fd931bc9a9f51b38))
* Port 'implement container restart API' patch ([5d9b1e0](https://github.com/Shubhranshu153/finch-daemon/commit/5d9b1e0f4e1565fd374b0f0941f373a094dc749c))


### Bug Fixes

* Add arm64 build and release ([6c622d7](https://github.com/Shubhranshu153/finch-daemon/commit/6c622d73de54e84b2fdd458f15c67738d19089fc))
* Add static binaries to release ([#63](https://github.com/Shubhranshu153/finch-daemon/issues/63)) ([57a0c44](https://github.com/Shubhranshu153/finch-daemon/commit/57a0c44d56bbf0addbf5b8c78a2baebac61141ab))
* bump minor for feat ([65fbf5a](https://github.com/Shubhranshu153/finch-daemon/commit/65fbf5afaeb175d5660ff13acc639ec3d72ac273))
* doc nits and parameter casing ([#57](https://github.com/Shubhranshu153/finch-daemon/issues/57)) ([e22c156](https://github.com/Shubhranshu153/finch-daemon/commit/e22c156cc8bcb97f25c6f41a14e833203e8798ce))
* filter unsupported enable_icc option ([#36](https://github.com/Shubhranshu153/finch-daemon/issues/36)) ([6c5e72d](https://github.com/Shubhranshu153/finch-daemon/commit/6c5e72d4e8c9f6a5be12bf38078798423d11064f))
* image load should close stream after copy ([#34](https://github.com/Shubhranshu153/finch-daemon/issues/34)) ([5ee657b](https://github.com/Shubhranshu153/finch-daemon/commit/5ee657b17de96c1d2302e9ee7490ccfdc64cd907))
* README changes re: systemd setup ([#59](https://github.com/Shubhranshu153/finch-daemon/issues/59)) ([2096ded](https://github.com/Shubhranshu153/finch-daemon/commit/2096ded2283a8582186be01eeee42a8c0ab6161d))
* Remove bump-patch-for-minor-pre-major ([#83](https://github.com/Shubhranshu153/finch-daemon/issues/83)) ([65fbf5a](https://github.com/Shubhranshu153/finch-daemon/commit/65fbf5afaeb175d5660ff13acc639ec3d72ac273))
* Set release version to 0.9.0 ([#56](https://github.com/Shubhranshu153/finch-daemon/issues/56)) ([024768a](https://github.com/Shubhranshu153/finch-daemon/commit/024768a6937ab2917870f9a3348dc0be114d3523))
* truncate image id on publish tag event ([#35](https://github.com/Shubhranshu153/finch-daemon/issues/35)) ([6aa5b7c](https://github.com/Shubhranshu153/finch-daemon/commit/6aa5b7ce76979682ad1cf2b49ac0237a74cac809))

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
