# Changelog

## [0.19.1](https://github.com/runfinch/finch-daemon/compare/v0.19.0...v0.19.1) (2025-07-28)


### Bug Fixes

* downgrade nerdctl from v2.1.3 to v2.1.2 ([#289](https://github.com/runfinch/finch-daemon/issues/289)) ([a7487bf](https://github.com/runfinch/finch-daemon/commit/a7487bfd915424f7e643426f3e60614e347162f6))

## [0.19.0](https://github.com/runfinch/finch-daemon/compare/v0.18.1...v0.19.0) (2025-07-17)


### Features

* Use nerdctl parsing logic for port publishing ([#265](https://github.com/runfinch/finch-daemon/issues/265)) ([1aec7ce](https://github.com/runfinch/finch-daemon/commit/1aec7cefe590a9f7d7857615f6ebebacd887545a))


### Bug Fixes

* restore update network settings for dockercompat ([#286](https://github.com/runfinch/finch-daemon/issues/286)) ([8fc3a1e](https://github.com/runfinch/finch-daemon/commit/8fc3a1eccf2ac5bc880dd184d8c13f58996836e7))

## [0.18.1](https://github.com/runfinch/finch-daemon/compare/v0.18.0...v0.18.1) (2025-07-11)


### Bug Fixes

* verify release artifact ([20dc067](https://github.com/runfinch/finch-daemon/commit/20dc0677673c88f2bacbe7b7f94d307899918108))
* verify release artifact docker-credential-finch ([#284](https://github.com/runfinch/finch-daemon/issues/284)) ([20dc067](https://github.com/runfinch/finch-daemon/commit/20dc0677673c88f2bacbe7b7f94d307899918108))

## [0.18.0](https://github.com/runfinch/finch-daemon/compare/v0.17.2...v0.18.0) (2025-07-11)


### Build System or External Dependencies

* **deps:** Bump github.com/docker/docker from 28.0.2+incompatible to 28.2.2+incompatible ([#251](https://github.com/runfinch/finch-daemon/issues/251)) ([097fb7e](https://github.com/runfinch/finch-daemon/commit/097fb7ee7badd55f4d4cecd671eaaad290c6e92f))
* **deps:** Bump github.com/go-viper/mapstructure/v2 ([0e28455](https://github.com/runfinch/finch-daemon/commit/0e2845588b0b5d922e539359f968c73fd271ac14))
* **deps:** Bump github.com/go-viper/mapstructure/v2 from 2.2.1 to 2.3.0 ([#273](https://github.com/runfinch/finch-daemon/issues/273)) ([0e28455](https://github.com/runfinch/finch-daemon/commit/0e2845588b0b5d922e539359f968c73fd271ac14))
* **deps:** Bump github.com/open-policy-agent/opa from 1.1.0 to 1.4.0 ([#268](https://github.com/runfinch/finch-daemon/issues/268)) ([7ec7d02](https://github.com/runfinch/finch-daemon/commit/7ec7d025266aa4cca0e137ee91390a16d1a21330))


### Features

* Add credential management for container build ([#275](https://github.com/runfinch/finch-daemon/issues/275)) ([47d2a65](https://github.com/runfinch/finch-daemon/commit/47d2a6560e6f5864b6aa4f4030e5ad8ebf977c5e))
* migrate from golang gomock to uber gomock ([#264](https://github.com/runfinch/finch-daemon/issues/264)) ([bb9442a](https://github.com/runfinch/finch-daemon/commit/bb9442a022aeb392822a697f13a238b7f81b8af8))
* Opa middleware support (Experimental) ([#156](https://github.com/runfinch/finch-daemon/issues/156)) ([91b9ac6](https://github.com/runfinch/finch-daemon/commit/91b9ac673ff13bcbe2a948d953481f5505245c4c))
  

## [0.17.2](https://github.com/runfinch/finch-daemon/compare/v0.17.1...v0.17.2) (2025-06-06)


### Build System or External Dependencies

* **deps:** Bump github.com/containerd/containerd/api ([439fd61](https://github.com/runfinch/finch-daemon/commit/439fd6185fbc9665f848ad0e8bade35de96ed001))
* **deps:** Bump github.com/containerd/containerd/api from 1.8.0 to 1.9.0 ([#245](https://github.com/runfinch/finch-daemon/issues/245)) ([439fd61](https://github.com/runfinch/finch-daemon/commit/439fd6185fbc9665f848ad0e8bade35de96ed001))
* **deps:** Bump github.com/docker/cli ([350ae05](https://github.com/runfinch/finch-daemon/commit/350ae05c60e69e823dfef8b9ba305d621f8eae5f))
* **deps:** Bump github.com/docker/cli from 28.0.4+incompatible to 28.2.2+incompatible ([#248](https://github.com/runfinch/finch-daemon/issues/248)) ([350ae05](https://github.com/runfinch/finch-daemon/commit/350ae05c60e69e823dfef8b9ba305d621f8eae5f))


### Bug Fixes

* skip blkio test in vm ([#256](https://github.com/runfinch/finch-daemon/issues/256)) ([7020f93](https://github.com/runfinch/finch-daemon/commit/7020f93040ff10ddf2d2724e439fea1fa7131381))

## [0.17.1](https://github.com/runfinch/finch-daemon/compare/v0.17.0...v0.17.1) (2025-06-06)


### Bug Fixes

* Update blkio tests to be compatible with finch CI ([#252](https://github.com/runfinch/finch-daemon/issues/252)) ([1d672cd](https://github.com/runfinch/finch-daemon/commit/1d672cd606476a7a6f05112c1b0dbb55959b55ad))

## [0.17.0](https://github.com/runfinch/finch-daemon/compare/v0.16.0...v0.17.0) (2025-06-04)


### Build System or External Dependencies

* **deps:** Bump github.com/containernetworking/cni from 1.2.3 to 1.3.0 ([#214](https://github.com/runfinch/finch-daemon/issues/214)) ([253b7eb](https://github.com/runfinch/finch-daemon/commit/253b7eb60bd4fa89cb96a1d043491a36a132b314))
* **deps:** Bump golang.org/x/sys from 0.32.0 to 0.33.0 ([#242](https://github.com/runfinch/finch-daemon/issues/242)) ([e706101](https://github.com/runfinch/finch-daemon/commit/e70610173485d724af390c2cc01b985681dab0e5))


### Features

* Add annotations option ([#238](https://github.com/runfinch/finch-daemon/issues/238)) ([23fd05e](https://github.com/runfinch/finch-daemon/commit/23fd05e60514217f19363cc41dd9db0a187b4b2f))
* Add Blkio related options ([#229](https://github.com/runfinch/finch-daemon/issues/229)) ([8dc97f8](https://github.com/runfinch/finch-daemon/commit/8dc97f832df7eabc68f3191a0c718bfbcfc7dde9))
* Add cgroupnsmode option ([#237](https://github.com/runfinch/finch-daemon/issues/237)) ([831e2a2](https://github.com/runfinch/finch-daemon/commit/831e2a2fc9405ace88d272d83a91602eb323bf40))
* Add devices option ([#236](https://github.com/runfinch/finch-daemon/issues/236)) ([4198dcc](https://github.com/runfinch/finch-daemon/commit/4198dccd88352cefef813302697b7f6606e6d869))
* Add PidMode, Ipc mode ([dde482c](https://github.com/runfinch/finch-daemon/commit/dde482cf96fc7d4f80d094c6e35b5cffd134fb1e))
* Add PidMode, Ipcmode and GroupAdd options ([#232](https://github.com/runfinch/finch-daemon/issues/232)) ([dde482c](https://github.com/runfinch/finch-daemon/commit/dde482cf96fc7d4f80d094c6e35b5cffd134fb1e))
* Add ReadonlyRootfs option ([#233](https://github.com/runfinch/finch-daemon/issues/233)) ([82b0ff4](https://github.com/runfinch/finch-daemon/commit/82b0ff425d97aeb1174965b8d22f3e67b4c79fbc))
* Add ShmSize, Sysctl and Runtime option ([#235](https://github.com/runfinch/finch-daemon/issues/235)) ([c4fc1c9](https://github.com/runfinch/finch-daemon/commit/c4fc1c951a42d22653f61e167324cc967f3d1901))
* add signal option to containerStop ([#158](https://github.com/runfinch/finch-daemon/issues/158)) ([abfa7f7](https://github.com/runfinch/finch-daemon/commit/abfa7f726cdb7f8a4ae45ae8ba4519b14805cf03))
* Add VolumesFrom, Tmpfs and UTSMode option ([#231](https://github.com/runfinch/finch-daemon/issues/231)) ([18434d7](https://github.com/runfinch/finch-daemon/commit/18434d7781d932efed1c937f861d7acba48f8c74))


### Bug Fixes

* handleNetworkLabels failing after nerdctl v2.0.0 upgrade ([#249](https://github.com/runfinch/finch-daemon/issues/249)) ([06bc4fd](https://github.com/runfinch/finch-daemon/commit/06bc4fdfbbf7e06361dd4449f8339db547657665))

## [0.16.0](https://github.com/runfinch/finch-daemon/compare/v0.15.0...v0.16.0) (2025-05-13)


### Build System or External Dependencies

* **deps:** Bump github.com/moby/moby ([c683e9c](https://github.com/runfinch/finch-daemon/commit/c683e9cf2419d9c3932bc1ab4f1e5bf78db8e95a))
* **deps:** Bump github.com/moby/moby from 28.0.1+incompatible to 28.1.1+incompatible ([#221](https://github.com/runfinch/finch-daemon/issues/221)) ([c683e9c](https://github.com/runfinch/finch-daemon/commit/c683e9cf2419d9c3932bc1ab4f1e5bf78db8e95a))
* **deps:** Bump github.com/onsi/gomega from 1.36.3 to 1.37.0 ([#216](https://github.com/runfinch/finch-daemon/issues/216)) ([c1b516e](https://github.com/runfinch/finch-daemon/commit/c1b516e0208c62225733dea8b523c501ba984e74))
* **deps:** Bump github.com/pelletier/go-toml/v2 from 2.2.3 to 2.2.4 ([#215](https://github.com/runfinch/finch-daemon/issues/215)) ([6c55ee6](https://github.com/runfinch/finch-daemon/commit/6c55ee611606ce8ed258ada429643202e3950c6e))
* **deps:** Bump github.com/runfinch/common-tests from 0.9.2 to 0.9.4 ([#239](https://github.com/runfinch/finch-daemon/issues/239)) ([3cfd4a0](https://github.com/runfinch/finch-daemon/commit/3cfd4a0f84cf0b1a14e050fec8adc9439bc7649f))
* **deps:** Bump google.golang.org/protobuf from 1.36.5 to 1.36.6 ([#213](https://github.com/runfinch/finch-daemon/issues/213)) ([8943984](https://github.com/runfinch/finch-daemon/commit/8943984c7cec5a3b208a960f4eb3e6e112771d77))


### Features

* Add Cpu options and missing tests to cidfile option ([#230](https://github.com/runfinch/finch-daemon/issues/230)) ([669a0ee](https://github.com/runfinch/finch-daemon/commit/669a0eeb974b8c72c884eb01503228e0842552f3))
* Add OomKillDisabled, NetworkDisabled and MACAddress option ([#228](https://github.com/runfinch/finch-daemon/issues/228)) ([58eef04](https://github.com/runfinch/finch-daemon/commit/58eef04d66523c9c994a6194d709fbce31cc5cbc))
* Add unpause container support ([#192](https://github.com/runfinch/finch-daemon/issues/192)) ([460b73f](https://github.com/runfinch/finch-daemon/commit/460b73f5872c7579be20572dde98bfd69ee9af14))

## [0.15.0](https://github.com/runfinch/finch-daemon/compare/v0.14.0...v0.15.0) (2025-04-10)


### Build System or External Dependencies

* **deps:** Bump github.com/containerd/nerdctl/v2 from 2.0.3 to 2.0.4 ([#207](https://github.com/runfinch/finch-daemon/issues/207)) ([f285fcd](https://github.com/runfinch/finch-daemon/commit/f285fcdc20cabdd2eda0089360e9be31d9e4ee36))
* **deps:** Bump github.com/docker/cli ([a00d384](https://github.com/runfinch/finch-daemon/commit/a00d384bc60cfe6aeeb5f2476febd24c6154db16))
* **deps:** Bump github.com/docker/cli from 27.5.0+incompatible to 28.0.4+incompatible ([#202](https://github.com/runfinch/finch-daemon/issues/202)) ([a00d384](https://github.com/runfinch/finch-daemon/commit/a00d384bc60cfe6aeeb5f2476febd24c6154db16))
* **deps:** Bump github.com/onsi/ginkgo/v2 from 2.22.2 to 2.23.4 ([#209](https://github.com/runfinch/finch-daemon/issues/209)) ([9ef5131](https://github.com/runfinch/finch-daemon/commit/9ef513179c764717aadc75a43c5312ac85241c11))
* **deps:** Bump github.com/opencontainers/runtime-spec ([9d4de73](https://github.com/runfinch/finch-daemon/commit/9d4de739318280715e4d697fd4d95e7fe576593b))
* **deps:** Bump github.com/opencontainers/runtime-spec from 1.2.0 to 1.2.1 ([#188](https://github.com/runfinch/finch-daemon/issues/188)) ([9d4de73](https://github.com/runfinch/finch-daemon/commit/9d4de739318280715e4d697fd4d95e7fe576593b))
* **deps:** Bump github.com/runfinch/common-tests from 0.9.1 to 0.9.2 ([#205](https://github.com/runfinch/finch-daemon/issues/205)) ([e5cef2a](https://github.com/runfinch/finch-daemon/commit/e5cef2a45900a141b3cdd0583709bde714af6642))
* **deps:** Bump github.com/spf13/afero from 1.12.0 to 1.14.0 ([#191](https://github.com/runfinch/finch-daemon/issues/191)) ([3dffcbb](https://github.com/runfinch/finch-daemon/commit/3dffcbbc8006bf12a65aa7ca6b87c5a3ed2dca6b))
* **deps:** Bump github.com/spf13/cobra from 1.8.1 to 1.9.1 ([#189](https://github.com/runfinch/finch-daemon/issues/189)) ([06b6922](https://github.com/runfinch/finch-daemon/commit/06b6922e8690142515ca5f3765ec03d0caba9d4c))
* **deps:** Bump golang.org/x/net from 0.37.0 to 0.39.0 ([#210](https://github.com/runfinch/finch-daemon/issues/210)) ([e22380f](https://github.com/runfinch/finch-daemon/commit/e22380f102921fe9ff143c4d91b6bf84fda80ef6))


### Features

* add additional options to image build ([#152](https://github.com/runfinch/finch-daemon/issues/152)) ([49cdd07](https://github.com/runfinch/finch-daemon/commit/49cdd075edcc342875b6fc6f3cb7b8fef4564c1f))
* Add pause container support ([#185](https://github.com/runfinch/finch-daemon/issues/185)) ([9aa41bf](https://github.com/runfinch/finch-daemon/commit/9aa41bfc0a82b869b2001c1e9c8aaf27b9a51959))


### Bug Fixes

* Tag mockgen and stringer packages for gen-code make target ([#193](https://github.com/runfinch/finch-daemon/issues/193)) ([3db64f9](https://github.com/runfinch/finch-daemon/commit/3db64f91cf23de2f1066cc1088c14fb7a1acb2ff))

## [0.14.0](https://github.com/runfinch/finch-daemon/compare/v0.13.1...v0.14.0) (2025-03-18)


### Build System or External Dependencies

* **deps:** Bump github.com/containerd/containerd/v2 from 2.0.2 to 2.0.4 ([#186](https://github.com/runfinch/finch-daemon/issues/186)) ([8d92e1f](https://github.com/runfinch/finch-daemon/commit/8d92e1f2a1d8555f4db1ed41025f7f9cac0916d0))
* **deps:** Bump github.com/containerd/go-cni from 1.1.11 to 1.1.12 ([#154](https://github.com/runfinch/finch-daemon/issues/154)) ([d9be44f](https://github.com/runfinch/finch-daemon/commit/d9be44f30aa0ae63aa1dd2b8224ff43eb8b11e0f))
* **deps:** bump github.com/containerd/nerdctl/v2 from 2.0.0 to 2.0.3 ([#173](https://github.com/runfinch/finch-daemon/issues/173)) ([1bcfebd](https://github.com/runfinch/finch-daemon/commit/1bcfebd4acc8c14c91eeda3c1e73b0e66a14b66e))
* **deps:** Bump github.com/docker/docker from 27.5.0+incompatible to 28.0.1+incompatible ([#170](https://github.com/runfinch/finch-daemon/issues/170)) ([e474488](https://github.com/runfinch/finch-daemon/commit/e474488f13615308af0e12f4a45720cfd947ba31))
* **deps:** Bump github.com/go-jose/go-jose/v4 from 4.0.4 to 4.0.5 ([#167](https://github.com/runfinch/finch-daemon/issues/167)) ([71a062a](https://github.com/runfinch/finch-daemon/commit/71a062adc537fa935e88e2da4c2b7addf4365c88))
* **deps:** Bump github.com/moby/moby from 27.4.1+incompatible to 28.0.1+incompatible ([#169](https://github.com/runfinch/finch-daemon/issues/169)) ([be4b61a](https://github.com/runfinch/finch-daemon/commit/be4b61a21cf68e0e67c29b830f6a09a5a11b4435))
* **deps:** Bump golang.org/x/net from 0.34.0 to 0.36.0 ([#182](https://github.com/runfinch/finch-daemon/issues/182)) ([006c68d](https://github.com/runfinch/finch-daemon/commit/006c68d3d55b6f2cd2e43f9f177c4ad503132f1b))
* **deps:** Bump golang.org/x/net from 0.34.0 to 0.37.0 ([#183](https://github.com/runfinch/finch-daemon/issues/183)) ([159c5f3](https://github.com/runfinch/finch-daemon/commit/159c5f3b860746f7b3e157664c10a0f6e238df0e))
* **deps:** Bump google.golang.org/protobuf from 1.36.4 to 1.36.5 ([#184](https://github.com/runfinch/finch-daemon/issues/184)) ([46b1402](https://github.com/runfinch/finch-daemon/commit/46b1402e070c3e662ca26fbf9710c04799d21fc2))


### Features

* add detachKeys option to container start ([#159](https://github.com/runfinch/finch-daemon/issues/159)) ([b03f126](https://github.com/runfinch/finch-daemon/commit/b03f12688a0bbbbbcad4f6a6ec300a39846775d1))


### Bug Fixes

* Refactor filters parsing ([#181](https://github.com/runfinch/finch-daemon/issues/181)) ([6d36c03](https://github.com/runfinch/finch-daemon/commit/6d36c03b37e5f1a7eb225093aff71e16728a5247))
* refactor wait API ([#177](https://github.com/runfinch/finch-daemon/issues/177)) ([08878dc](https://github.com/runfinch/finch-daemon/commit/08878dc134e24310c293849950e854c83ff30cb5))

## [0.13.1](https://github.com/runfinch/finch-daemon/compare/v0.13.0...v0.13.1) (2025-02-21)


### Bug Fixes

* update create-releases.sh ([#165](https://github.com/runfinch/finch-daemon/issues/165)) ([1edf2d4](https://github.com/runfinch/finch-daemon/commit/1edf2d430e1f9e7c682ae8bd85347b7b97a52e75))

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
