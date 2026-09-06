# Changelog

## [0.3.1](https://github.com/ishioni/eep/compare/0.3.0...0.3.1) (2026-09-06)


### Bug Fixes

* **container:** update image envoyproxy/envoy (v1.39.0 → v1.39.1) ([#35](https://github.com/ishioni/eep/issues/35)) ([ca9a5fb](https://github.com/ishioni/eep/commit/ca9a5fb4515141f4e840c131099a539b57d59f13))
* **mise:** update tool go (1.27.0 → 1.27.1) ([#39](https://github.com/ishioni/eep/issues/39)) ([9e2dc26](https://github.com/ishioni/eep/commit/9e2dc264b5324b4cf275fa035d0a44838b38dd64))


### Miscellaneous Chores

* **github-action:** update action docker/github-builder (v1.16.0 → v1.17.0) ([#32](https://github.com/ishioni/eep/issues/32)) ([2f8a730](https://github.com/ishioni/eep/commit/2f8a730affba531696bc9825121074acfd5a8338))
* **github-action:** update github-actions ([#34](https://github.com/ishioni/eep/issues/34)) ([a642523](https://github.com/ishioni/eep/commit/a6425238672a7462c668b5118c31d61a3e630ae2))
* **mise:** update mise tools ([#38](https://github.com/ishioni/eep/issues/38)) ([793112d](https://github.com/ishioni/eep/commit/793112d90dd00bd465e2ed4ada9407be7248c690))
* **mise:** update tool node (24.19.0 → v24.20.0) ([#37](https://github.com/ishioni/eep/issues/37)) ([9f79d1a](https://github.com/ishioni/eep/commit/9f79d1a62404e7c70cd02763eb369bb4427fc136))
* **mise:** update tool oxfmt (0.64.0 → 0.65.0) ([#36](https://github.com/ishioni/eep/issues/36)) ([1d075eb](https://github.com/ishioni/eep/commit/1d075eb49e1c73a7058f4742d90edc946d23239e))
* **mise:** update tool oxfmt (0.65.0 → 0.66.0) ([#40](https://github.com/ishioni/eep/issues/40)) ([0d26a49](https://github.com/ishioni/eep/commit/0d26a4903255c4775ff7a08489cd68b1d673c9e5))

## [0.3.0](https://github.com/ishioni/eep/compare/0.2.1...0.3.0) (2026-08-22)


### Features

* **mise:** update tool go (1.26.6 → 1.27.0) ([#30](https://github.com/ishioni/eep/issues/30)) ([3007fbd](https://github.com/ishioni/eep/commit/3007fbd88cb698a0677fa8a15c4929074f49174e))


### Bug Fixes

* **ci:** add node ([d8c6ca8](https://github.com/ishioni/eep/commit/d8c6ca87923078301116966b937ff30e9b8afd11))
* **ci:** bump go.mod together with go version in mise ([9628bd5](https://github.com/ishioni/eep/commit/9628bd585f12c4cebb2c503cfc0be9f139a15ea2))


### Miscellaneous Chores

* **envoy-gateway:** add http2.initialStreamWindowSize to example ctp ([#25](https://github.com/ishioni/eep/issues/25)) ([01eb774](https://github.com/ishioni/eep/commit/01eb774e8f43db887df6415973f8a385826c5851))
* **github-action:** update action docker/setup-buildx-action (v4.2.0 → v4.3.0) ([#28](https://github.com/ishioni/eep/issues/28)) ([bf064bf](https://github.com/ishioni/eep/commit/bf064bf43c26a7e0c83364043b16a3f0fe3fd235))
* **mise:** update mise tools ([#29](https://github.com/ishioni/eep/issues/29)) ([91950d9](https://github.com/ishioni/eep/commit/91950d916efaa1bb568c32b4ca414f9adc680f67))
* **mise:** update tool oxfmt (0.63.0 → 0.64.0) ([#27](https://github.com/ishioni/eep/issues/27)) ([72cb83d](https://github.com/ishioni/eep/commit/72cb83dc0a4d3c031e74b17a8adcea36c61227bc))
* run go fix with Go 1.27 ([#31](https://github.com/ishioni/eep/issues/31)) ([d3d6f2c](https://github.com/ishioni/eep/commit/d3d6f2c549e2931c1054d39322746fad33322fb8))

## [0.2.1](https://github.com/ishioni/eep/compare/0.2.0...0.2.1) (2026-08-17)


### Bug Fixes

* validate plugin configuration fields ([#23](https://github.com/ishioni/eep/issues/23)) ([34d2d8d](https://github.com/ishioni/eep/commit/34d2d8d11fa62c47b6fe9ee8fe9a46eb5626c0bb))


### Documentation

* configure Envoy Gateway response buffering ([#24](https://github.com/ishioni/eep/issues/24)) ([42ea3ae](https://github.com/ishioni/eep/commit/42ea3ae79105595f235835388ccc1839141314b3))
* fix examples ([554f1b4](https://github.com/ishioni/eep/commit/554f1b4f2630324714832db0aa577e98b85f9f84))

## [0.2.0](https://github.com/ishioni/eep/compare/0.1.0...0.2.0) (2026-08-16)


### Features

* add configurable log level ([#21](https://github.com/ishioni/eep/issues/21)) ([169415a](https://github.com/ishioni/eep/commit/169415afec0253a9177af3364ee4e90a2ee2105f))
* add response and domain filters ([#19](https://github.com/ishioni/eep/issues/19)) ([ee20a0f](https://github.com/ishioni/eep/commit/ee20a0fc3c9b1a7a991b87cc2765047b579139f1))


### Bug Fixes

* **release:** bump features before 1.0 ([#20](https://github.com/ishioni/eep/issues/20)) ([f4a28c2](https://github.com/ishioni/eep/commit/f4a28c20485511996ff44fa6b058a5381fc1b602))


### Documentation

* acknowledge error-pages MIT license ([d68f024](https://github.com/ishioni/eep/commit/d68f0247e6a270f42383e4740905e4e3e6a416bf))


### Continuous Integration

* publish a single wasm transport image ([#17](https://github.com/ishioni/eep/issues/17)) ([d2ae207](https://github.com/ishioni/eep/commit/d2ae20748543b0cea8d136e2554124c55ee8fa9f))

## 0.1.0 (2026-08-16)


### Features

* add localized error pages ([#11](https://github.com/ishioni/eep/issues/11)) ([b744728](https://github.com/ishioni/eep/commit/b744728905aa9a2abbf438d225386339edf6385d))
* add runtime plugin configuration ([#9](https://github.com/ishioni/eep/issues/9)) ([b062683](https://github.com/ishioni/eep/commit/b062683e3d8bde3eb0dd9f7e1aabe29f42384ec6))
* **container:** update image envoyproxy/envoy (v1.38.0 → v1.39.0) ([#5](https://github.com/ishioni/eep/issues/5)) ([03650df](https://github.com/ishioni/eep/commit/03650df105abdcc0618210aa0c8a7863e47c9ffe))
* negotiate non-HTML error responses ([#7](https://github.com/ishioni/eep/issues/7)) ([0b5a419](https://github.com/ishioni/eep/commit/0b5a41992ce142f437e4f1444db3ac2a8b6ccc08))


### Bug Fixes

* disable request details by default ([9fdc57d](https://github.com/ishioni/eep/commit/9fdc57d247b209fabab785367f19f85b94b9436a))


### Code Refactoring

* isolate error page rendering ([#10](https://github.com/ishioni/eep/issues/10)) ([44ecc13](https://github.com/ishioni/eep/commit/44ecc131d136b50eb8c482fab7d7398a42368d0a))
* make Envoy smoke test self-contained ([#12](https://github.com/ishioni/eep/issues/12)) ([992c73f](https://github.com/ishioni/eep/commit/992c73fd19846929d047d1315aad5cafa8b2b2b1))
* move wasm entrypoint into cmd/eep ([#13](https://github.com/ishioni/eep/issues/13)) ([1f1f2a8](https://github.com/ishioni/eep/commit/1f1f2a8394d50f269ddecb205af680a795a35547))
* rename project to eep ([208c1b4](https://github.com/ishioni/eep/commit/208c1b4cf5f2710ce5296ab06cca012f9deecb03))
* require Envoy 1.39 ([#8](https://github.com/ishioni/eep/issues/8)) ([a4ed346](https://github.com/ishioni/eep/commit/a4ed346b0dea2f1f8b3e26c08df518e3866f7637))


### Documentation

* consolidate quickstart into readme ([#15](https://github.com/ishioni/eep/issues/15)) ([aba9509](https://github.com/ishioni/eep/commit/aba9509d113c79fca5326c0c022d70a7bf3f6967))
* restructure readme for release ([#16](https://github.com/ishioni/eep/issues/16)) ([ee6105a](https://github.com/ishioni/eep/commit/ee6105a06cc88a1890ffe1bc5b145b4e304de405))


### Build System

* modernize project tooling ([#2](https://github.com/ishioni/eep/issues/2)) ([d174c84](https://github.com/ishioni/eep/commit/d174c84e24456b4dc7e44f6bb4bf5f71a80c8ea3))


### Miscellaneous Chores

* **mise:** update tool go:golang.org/x/vuln/cmd/govulncheck (1.6.0 → v1.7.0) ([#4](https://github.com/ishioni/eep/issues/4)) ([f42e30a](https://github.com/ishioni/eep/commit/f42e30adde677e9f1d0aff588c23cd9b0eb78f34))
* remove obsolete Compose builder ([c6d0f6c](https://github.com/ishioni/eep/commit/c6d0f6cafe02d02509723c4a668fd34f6fded464))

## Changelog

All notable changes to this project will be documented in this file.
