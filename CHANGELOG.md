## [3.5.1](https://github.com/jarv/newsgoat/compare/v3.5.0...v3.5.1) (2026-04-27)


### Bug Fixes

* apply filter to unread counts in feed list view ([23017e0](https://github.com/jarv/newsgoat/commit/23017e00a4befbf78cf330a3dad474d2475a8c58))

# [3.5.0](https://github.com/jarv/newsgoat/compare/v3.4.2...v3.5.0) (2026-04-27)


### Features

* add OPML import command ([#138](https://github.com/jarv/newsgoat/issues/138)) ([dbbc9fd](https://github.com/jarv/newsgoat/commit/dbbc9fdbe299c252a839426d5fd57a592befe3bd))

## [3.4.2](https://github.com/jarv/newsgoat/compare/v3.4.1...v3.4.2) (2026-04-27)


### Bug Fixes

* remove filter count coloring, keep only ✦ indicator for configured filters ([a3a2b50](https://github.com/jarv/newsgoat/commit/a3a2b50315df10085764b61bff9e144c9403c8d2))

## [3.4.1](https://github.com/jarv/newsgoat/compare/v3.4.0...v3.4.1) (2026-04-27)


### Bug Fixes

* only show ✦ on the item where the filter is directly set ([caabcf4](https://github.com/jarv/newsgoat/commit/caabcf4a91a25b2d723c5cc0b6ff0f8a5c242f00))

# [3.4.0](https://github.com/jarv/newsgoat/compare/v3.3.2...v3.4.0) (2026-04-27)


### Features

* add ✦ indicator for feeds/folders with a configured filter ([344f752](https://github.com/jarv/newsgoat/commit/344f752fb49f24497fd303fa465dcc4b9be2e236))

## [3.3.2](https://github.com/jarv/newsgoat/compare/v3.3.1...v3.3.2) (2026-04-27)


### Bug Fixes

* only color feed unread count when filter actually changed it ([6d4a290](https://github.com/jarv/newsgoat/commit/6d4a290d708bfda0b074429db008ade60f30c779))

## [3.3.1](https://github.com/jarv/newsgoat/compare/v3.3.0...v3.3.1) (2026-04-27)


### Bug Fixes

* keep unread style (blue) based on original count, only color the count orange when filtered ([7321356](https://github.com/jarv/newsgoat/commit/73213566576abcc2952871c3402995c82082066e))
* render filtered count orange inline without breaking unread blue style ([b563a71](https://github.com/jarv/newsgoat/commit/b563a71c2ee5097d5089e86f28beddf71d1ff326))
* use title bar color for filtered unread count instead of orange ([9c8a10a](https://github.com/jarv/newsgoat/commit/9c8a10a341a1ff8c95ed4e6b0b7477151ce995e3))

# [3.3.0](https://github.com/jarv/newsgoat/compare/v3.2.1...v3.3.0) (2026-04-27)


### Features

* add regex-based filters for feeds and folders ([#137](https://github.com/jarv/newsgoat/issues/137)) ([3ad29ca](https://github.com/jarv/newsgoat/commit/3ad29ca218636617b7607a204a3cf5400ef8a55d))

## [3.2.1](https://github.com/jarv/newsgoat/compare/v3.2.0...v3.2.1) (2026-04-27)


### Bug Fixes

* configure git identity in homebrew publish script for CI ([b4e4402](https://github.com/jarv/newsgoat/commit/b4e44022b1669e55467525555e71526f05c2c28e))

# [3.2.0](https://github.com/jarv/newsgoat/compare/v3.1.0...v3.2.0) (2026-04-27)


### Features

* add Homebrew tap support for macOS installation ([#136](https://github.com/jarv/newsgoat/issues/136)) ([2095949](https://github.com/jarv/newsgoat/commit/2095949ce6addb124dc61731cb0be97786bc5400)), closes [#88](https://github.com/jarv/newsgoat/issues/88)


### Reverts

* remove version check on launch feature (a70a2e9) ([#135](https://github.com/jarv/newsgoat/issues/135)) ([363ce31](https://github.com/jarv/newsgoat/commit/363ce31811e817feff4af3061c0f633010fbc6ca))

# [3.1.0](https://github.com/jarv/newsgoat/compare/v3.0.11...v3.1.0) (2026-04-27)


### Features

* upgrade charmbracelet bubbletea, lipgloss, glamour to v2 ([#133](https://github.com/jarv/newsgoat/issues/133)) ([592de6c](https://github.com/jarv/newsgoat/commit/592de6c83ab3446ff1861e9eee0c9589d8503e03))

## [3.0.11](https://github.com/jarv/newsgoat/compare/v3.0.10...v3.0.11) (2026-04-27)


### Bug Fixes

* **deps:** update module github.com/ncruces/go-sqlite3 to v0.34.0 ([#120](https://github.com/jarv/newsgoat/issues/120)) ([d325c43](https://github.com/jarv/newsgoat/commit/d325c43fb0a97ab9a405f016d79a61e18e14adf7))

## [3.0.10](https://github.com/jarv/newsgoat/compare/v3.0.9...v3.0.10) (2026-04-09)


### Bug Fixes

* **deps:** update module golang.org/x/net to v0.53.0 ([#128](https://github.com/jarv/newsgoat/issues/128)) ([feddde9](https://github.com/jarv/newsgoat/commit/feddde9c3045a0998f17b877c3047bb59d7dd651))

## [3.0.9](https://github.com/jarv/newsgoat/compare/v3.0.8...v3.0.9) (2026-03-13)


### Bug Fixes

* **deps:** update module github.com/ncruces/go-sqlite3 to v0.32.0 ([#118](https://github.com/jarv/newsgoat/issues/118)) ([11d562b](https://github.com/jarv/newsgoat/commit/11d562bdc24987325de807025334554dabfa64d3))

## [3.0.8](https://github.com/jarv/newsgoat/compare/v3.0.7...v3.0.8) (2026-03-12)


### Bug Fixes

* **deps:** update module github.com/ncruces/go-sqlite3 to v0.31.1 ([#116](https://github.com/jarv/newsgoat/issues/116)) ([dae3759](https://github.com/jarv/newsgoat/commit/dae37592d0b9b7d7c16894d20318a4da13f62afe))
* **deps:** update module golang.org/x/net to v0.52.0 ([#117](https://github.com/jarv/newsgoat/issues/117)) ([0d0674e](https://github.com/jarv/newsgoat/commit/0d0674e24821122a9b49852d464540b7a37eb5ec))

## [3.0.7](https://github.com/jarv/newsgoat/compare/v3.0.6...v3.0.7) (2026-03-10)


### Bug Fixes

* **deps:** update module github.com/ncruces/go-sqlite3 to v0.31.0 ([#114](https://github.com/jarv/newsgoat/issues/114)) ([cc77d76](https://github.com/jarv/newsgoat/commit/cc77d76578bb8059b634f876b0e26311eb340871))

## [3.0.6](https://github.com/jarv/newsgoat/compare/v3.0.5...v3.0.6) (2026-02-26)


### Bug Fixes

* **deps:** update module golang.org/x/net to v0.51.0 ([#105](https://github.com/jarv/newsgoat/issues/105)) ([c674219](https://github.com/jarv/newsgoat/commit/c674219e2d3c3b447c06fbfd73aeea86402d3935))

## [3.0.5](https://github.com/jarv/newsgoat/compare/v3.0.4...v3.0.5) (2026-02-09)


### Bug Fixes

* **deps:** update module golang.org/x/net to v0.50.0 ([#91](https://github.com/jarv/newsgoat/issues/91)) ([4e86920](https://github.com/jarv/newsgoat/commit/4e86920db627b972014c35d1b7c112ff09c82fc1))

## [3.0.4](https://github.com/jarv/newsgoat/compare/v3.0.3...v3.0.4) (2026-02-09)


### Bug Fixes

* **deps:** update github.com/charmbracelet/lipgloss digest to 1e3ee34 ([#90](https://github.com/jarv/newsgoat/issues/90)) ([2ca4bd0](https://github.com/jarv/newsgoat/commit/2ca4bd0a8eec44aaacd877eb4ebba0aecb0b384d))

## [3.0.3](https://github.com/jarv/newsgoat/compare/v3.0.2...v3.0.3) (2026-02-02)


### Bug Fixes

* **deps:** update github.com/charmbracelet/lipgloss digest to 832bc9d ([#85](https://github.com/jarv/newsgoat/issues/85)) ([3671178](https://github.com/jarv/newsgoat/commit/3671178cdccafc7cefecd71478bf79d8224c02dd))

## [3.0.2](https://github.com/jarv/newsgoat/compare/v3.0.1...v3.0.2) (2026-01-24)


### Bug Fixes

* **deps:** update module github.com/ncruces/go-sqlite3 to v0.30.5 ([#78](https://github.com/jarv/newsgoat/issues/78)) ([1d5bce2](https://github.com/jarv/newsgoat/commit/1d5bce233d1523ea82a5bc1982387c051f7c1bb3))

## [3.0.1](https://github.com/jarv/newsgoat/compare/v3.0.0...v3.0.1) (2026-01-12)


### Bug Fixes

* **deps:** update module golang.org/x/net to v0.49.0 ([#73](https://github.com/jarv/newsgoat/issues/73)) ([14abbb8](https://github.com/jarv/newsgoat/commit/14abbb83b8006cd71f24df7fd7a2108f29cbd321))

# [3.0.0](https://github.com/jarv/newsgoat/compare/v2.0.14...v3.0.0) (2026-01-08)


### Features

* Removes client/server mode ([a1f9d56](https://github.com/jarv/newsgoat/commit/a1f9d5629acb6dee948004e5a57423adf541450d))


### BREAKING CHANGES

* this reverts client/server implementation

The client/server implementation needs to be refactored, for now it will
be backed out.

## [2.0.14](https://github.com/jarv/newsgoat/compare/v2.0.13...v2.0.14) (2026-01-05)


### Bug Fixes

* **deps:** update github.com/charmbracelet/lipgloss digest to f2d1864 ([#70](https://github.com/jarv/newsgoat/issues/70)) ([59d4a5c](https://github.com/jarv/newsgoat/commit/59d4a5c231eecc8ef7e2a605917c6500ad27087b))

## [2.0.13](https://github.com/jarv/newsgoat/compare/v2.0.12...v2.0.13) (2025-12-23)


### Bug Fixes

* **deps:** update module google.golang.org/grpc to v1.78.0 ([#67](https://github.com/jarv/newsgoat/issues/67)) ([43ce41a](https://github.com/jarv/newsgoat/commit/43ce41a6bc8d722949031a2330dde7f6accaa71b))

## [2.0.12](https://github.com/jarv/newsgoat/compare/v2.0.11...v2.0.12) (2025-12-19)


### Bug Fixes

* **deps:** update module github.com/ncruces/go-sqlite3 to v0.30.4 ([#65](https://github.com/jarv/newsgoat/issues/65)) ([2b4111b](https://github.com/jarv/newsgoat/commit/2b4111b445625b2fca94e3ac0f9f7cd33203a2e9))

## [2.0.11](https://github.com/jarv/newsgoat/compare/v2.0.10...v2.0.11) (2025-12-12)


### Bug Fixes

* **deps:** update module google.golang.org/protobuf to v1.36.11 ([#62](https://github.com/jarv/newsgoat/issues/62)) ([ad4214f](https://github.com/jarv/newsgoat/commit/ad4214fc20585301aa2779e6e534ea3cf1d55450))

## [2.0.10](https://github.com/jarv/newsgoat/compare/v2.0.9...v2.0.10) (2025-12-08)


### Bug Fixes

* **deps:** update module golang.org/x/net to v0.48.0 ([#60](https://github.com/jarv/newsgoat/issues/60)) ([464c1cc](https://github.com/jarv/newsgoat/commit/464c1cc7fe991a199c09d1b0c5143daf23317834))

## [2.0.9](https://github.com/jarv/newsgoat/compare/v2.0.8...v2.0.9) (2025-12-02)


### Bug Fixes

* **deps:** update module github.com/ncruces/go-sqlite3 to v0.30.3 ([#54](https://github.com/jarv/newsgoat/issues/54)) ([ac96943](https://github.com/jarv/newsgoat/commit/ac96943de4b086bb9b70340392e920751632ab45))

## [2.0.8](https://github.com/jarv/newsgoat/compare/v2.0.7...v2.0.8) (2025-11-30)


### Bug Fixes

* **deps:** update module github.com/johanneskaufmann/html-to-markdown/v2 to v2.5.0 ([#52](https://github.com/jarv/newsgoat/issues/52)) ([efef1ce](https://github.com/jarv/newsgoat/commit/efef1ce0ccab257853fae39ab11ed1fb08cd2702))

## [2.0.7](https://github.com/jarv/newsgoat/compare/v2.0.6...v2.0.7) (2025-11-26)


### Bug Fixes

* **deps:** update module github.com/ncruces/go-sqlite3 to v0.30.2 ([#51](https://github.com/jarv/newsgoat/issues/51)) ([53970e3](https://github.com/jarv/newsgoat/commit/53970e3415902dcd942f1fc630ee4fbffda88815))

## [2.0.6](https://github.com/jarv/newsgoat/compare/v2.0.5...v2.0.6) (2025-11-24)


### Bug Fixes

* Fix remote log handling so we only see client logs in logging view ([a7f788b](https://github.com/jarv/newsgoat/commit/a7f788bc7ed825bd4ad549bedb01f041156d82cf))

## [2.0.5](https://github.com/jarv/newsgoat/compare/v2.0.4...v2.0.5) (2025-11-24)


### Bug Fixes

* **deps:** update github.com/charmbracelet/lipgloss digest to fabd8ab ([#50](https://github.com/jarv/newsgoat/issues/50)) ([e036f44](https://github.com/jarv/newsgoat/commit/e036f4477c650d087701fbaee5dccecda0639b4f))

## [2.0.4](https://github.com/jarv/newsgoat/compare/v2.0.3...v2.0.4) (2025-11-24)


### Bug Fixes

* Fixes grpc support for settings, config and folder ops ([463537c](https://github.com/jarv/newsgoat/commit/463537c8b5e9321bb77b00cf2d8c68be93c03250))

## [2.0.3](https://github.com/jarv/newsgoat/compare/v2.0.2...v2.0.3) (2025-11-24)


### Bug Fixes

* adds some info logging for server mode ([9ee2f5f](https://github.com/jarv/newsgoat/commit/9ee2f5f83377086d2559928c39c3541994380dfe))

## [2.0.2](https://github.com/jarv/newsgoat/compare/v2.0.1...v2.0.2) (2025-11-21)


### Bug Fixes

* **deps:** update module google.golang.org/grpc to v1.77.0 ([#48](https://github.com/jarv/newsgoat/issues/48)) ([4c83054](https://github.com/jarv/newsgoat/commit/4c83054544b0e76b70024b08e42b978b5092cbc5))

## [2.0.1](https://github.com/jarv/newsgoat/compare/v2.0.0...v2.0.1) (2025-11-21)


### Bug Fixes

* Increase task queue to 1000 to support more feeds ([6a3885a](https://github.com/jarv/newsgoat/commit/6a3885a16d38b0647578171bebbde3648e7dd53c))

# [2.0.0](https://github.com/jarv/newsgoat/compare/v1.11.14...v2.0.0) (2025-11-21)


### Features

* Client/server mode ([ebc1da7](https://github.com/jarv/newsgoat/commit/ebc1da7da2f6183001b981f75b169fbea62d2ab4))


### BREAKING CHANGES

* Removes the file-based url config

- client/server mode
- refactor logging

## [1.11.14](https://github.com/jarv/newsgoat/compare/v1.11.13...v1.11.14) (2025-11-11)


### Bug Fixes

* **deps:** update module golang.org/x/net to v0.47.0 ([#43](https://github.com/jarv/newsgoat/issues/43)) ([aff87c9](https://github.com/jarv/newsgoat/commit/aff87c973c58b53c769b780d2fe21c6ba19d6ba8))

## [1.11.13](https://github.com/jarv/newsgoat/compare/v1.11.12...v1.11.13) (2025-11-09)


### Bug Fixes

* **deps:** update module github.com/ncruces/go-sqlite3 to v0.30.1 ([#40](https://github.com/jarv/newsgoat/issues/40)) ([dcea75e](https://github.com/jarv/newsgoat/commit/dcea75e155dd479bb0f2f1186a04c48613f8ae1f))

## [1.11.12](https://github.com/jarv/newsgoat/compare/v1.11.11...v1.11.12) (2025-11-06)


### Bug Fixes

* Adds 'o' for open to help text ([dccae9f](https://github.com/jarv/newsgoat/commit/dccae9fec089bcc8d327c1393db332e5e314536c))

## [1.11.11](https://github.com/jarv/newsgoat/compare/v1.11.10...v1.11.11) (2025-11-05)


### Bug Fixes

* **deps:** update module github.com/ncruces/go-sqlite3 to v0.30.0 ([#35](https://github.com/jarv/newsgoat/issues/35)) ([f6a645d](https://github.com/jarv/newsgoat/commit/f6a645d3904e87f081eb8ccf767141f3859cbbbe))

## [1.11.10](https://github.com/jarv/newsgoat/compare/v1.11.9...v1.11.10) (2025-11-03)


### Bug Fixes

* **deps:** update github.com/charmbracelet/lipgloss digest to 7d1b622 ([#33](https://github.com/jarv/newsgoat/issues/33)) ([dd9aba5](https://github.com/jarv/newsgoat/commit/dd9aba5ede4c13dc6f788c8456768f258a9aeb34))

## [1.11.9](https://github.com/jarv/newsgoat/compare/v1.11.8...v1.11.9) (2025-10-31)


### Bug Fixes

* Adds feed title to the item list ([aec02af](https://github.com/jarv/newsgoat/commit/aec02afafa08674b7698d310ae145b9e44e43b91))

## [1.11.8](https://github.com/jarv/newsgoat/compare/v1.11.7...v1.11.8) (2025-10-30)


### Bug Fixes

* **deps:** update github.com/charmbracelet/lipgloss digest to 66093c8 ([#31](https://github.com/jarv/newsgoat/issues/31)) ([9c73f4f](https://github.com/jarv/newsgoat/commit/9c73f4f6956d94e88df62157edfd6d83876e3a87))

## [1.11.7](https://github.com/jarv/newsgoat/compare/v1.11.6...v1.11.7) (2025-10-20)


### Bug Fixes

* Formatting improvements to the article view ([f8a51ad](https://github.com/jarv/newsgoat/commit/f8a51ad34a0c808f5669f68a514238c9de985fb0))

## [1.11.6](https://github.com/jarv/newsgoat/compare/v1.11.5...v1.11.6) (2025-10-20)


### Bug Fixes

* Fixes non-html feed formatting (youtube) ([55703db](https://github.com/jarv/newsgoat/commit/55703dbd09f87a188fbf192ca6bc193721ea88ac))
* Reverts sort order for unread items on top for feed items ([bdab480](https://github.com/jarv/newsgoat/commit/bdab480dccf2b44c535e41f946c0187b862970a8))

## [1.11.5](https://github.com/jarv/newsgoat/compare/v1.11.4...v1.11.5) (2025-10-19)


### Bug Fixes

* Adds the unread indicator back but instead use 'N' ([400505a](https://github.com/jarv/newsgoat/commit/400505a8dea0545bcb974b4c995abafbf041a736))

## [1.11.4](https://github.com/jarv/newsgoat/compare/v1.11.3...v1.11.4) (2025-10-18)


### Bug Fixes

* Instead of a binary rename, copy the binary for updates ([eb7723b](https://github.com/jarv/newsgoat/commit/eb7723bad8e5940ec3c422580554ea5e4500dceb))

## [1.11.3](https://github.com/jarv/newsgoat/compare/v1.11.2...v1.11.3) (2025-10-18)


### Bug Fixes

* Use blue for the background selector for the Dracula theme ([7871b8d](https://github.com/jarv/newsgoat/commit/7871b8db0b5f22c27bc0bccadde7ea9cd9be8a3b))

## [1.11.2](https://github.com/jarv/newsgoat/compare/v1.11.1...v1.11.2) (2025-10-18)


### Bug Fixes

* Removes 🔵 for unread items ([7409982](https://github.com/jarv/newsgoat/commit/740998211ea4bed5899eed38359edd56b8dc96c0))

## [1.11.1](https://github.com/jarv/newsgoat/compare/v1.11.0...v1.11.1) (2025-10-17)


### Bug Fixes

* Extend unread feed sorting to the Feed Items view ([75046d8](https://github.com/jarv/newsgoat/commit/75046d87ca6b5acec34b3757598146b60b510627))
* Resolves permission error for the auto-updater ([af2934f](https://github.com/jarv/newsgoat/commit/af2934f2ee9217b386855bb128c0b4b0875b5e0f))

# [1.11.0](https://github.com/jarv/newsgoat/compare/v1.10.2...v1.11.0) (2025-10-17)


### Features

* Adds scrolling navigation for feed titles ([8ec9025](https://github.com/jarv/newsgoat/commit/8ec90251866142d827788bb3cb5d8faf2afaad04))

## [1.10.2](https://github.com/jarv/newsgoat/compare/v1.10.1...v1.10.2) (2025-10-16)


### Bug Fixes

* Improves explainer text for url file ([7367445](https://github.com/jarv/newsgoat/commit/73674454c8d98c2c1026a23cb31e8fdf7c724e83))

## [1.10.1](https://github.com/jarv/newsgoat/compare/v1.10.0...v1.10.1) (2025-10-16)


### Bug Fixes

* lint fixes and spurious log message ([107d875](https://github.com/jarv/newsgoat/commit/107d8750d52e44ed3afc72e4126319f464c560f0))

# [1.10.0](https://github.com/jarv/newsgoat/compare/v1.9.0...v1.10.0) (2025-10-16)


### Features

* Adds version check on start ([a70a2e9](https://github.com/jarv/newsgoat/commit/a70a2e911b9f309fc671cdb80c777ca622d8a0de))

# [1.9.0](https://github.com/jarv/newsgoat/compare/v1.8.2...v1.9.0) (2025-10-15)


### Bug Fixes

* Preserve comments in urls file when adding individual URLs ([58bcb61](https://github.com/jarv/newsgoat/commit/58bcb6195bd543bd8411ba29ef286998c5d43129))


### Features

* Improve feed search ([b1a4dd4](https://github.com/jarv/newsgoat/commit/b1a4dd43667e2736f35e49d4e0d45853d27a2a53))

## [1.8.2](https://github.com/jarv/newsgoat/compare/v1.8.1...v1.8.2) (2025-10-09)


### Bug Fixes

* Fixes broken RSS 1.0 feed parsing ([a502a7d](https://github.com/jarv/newsgoat/commit/a502a7d03e9bc5e202f6bbaf18925c41b349b7a3))

## [1.8.1](https://github.com/jarv/newsgoat/compare/v1.8.0...v1.8.1) (2025-10-09)


### Bug Fixes

* Resolves golangci-lint errors and updates go version ([842b0e4](https://github.com/jarv/newsgoat/commit/842b0e410ee67a83d9a258ccaff55742d20fdc9c))

# [1.8.0](https://github.com/jarv/newsgoat/compare/v1.7.3...v1.8.0) (2025-10-09)


### Features

* Adds folders for grouping feeds ([18ac04c](https://github.com/jarv/newsgoat/commit/18ac04c3efb5c24949958c93254950733663e3a7))
* Adds logic for db schema migrations ([4d94a18](https://github.com/jarv/newsgoat/commit/4d94a18b764bdbefccf5ed885802457c4980bbd5))

## [1.7.3](https://github.com/jarv/newsgoat/compare/v1.7.2...v1.7.3) (2025-10-08)


### Bug Fixes

* **deps:** update module golang.org/x/net to v0.46.0 ([#19](https://github.com/jarv/newsgoat/issues/19)) ([16c90c2](https://github.com/jarv/newsgoat/commit/16c90c2e64a552994ca8c071f2cb32e86054c555))

## [1.7.2](https://github.com/jarv/newsgoat/compare/v1.7.1...v1.7.2) (2025-10-07)


### Bug Fixes

* **deps:** update module golang.org/x/net to v0.45.0 ([#16](https://github.com/jarv/newsgoat/issues/16)) ([223bf5a](https://github.com/jarv/newsgoat/commit/223bf5a0cf099f230924177d4af21eb9d267a065))

## [1.7.1](https://github.com/jarv/newsgoat/compare/v1.7.0...v1.7.1) (2025-10-07)


### Bug Fixes

* Unread count was incorrect for feeds without any items ([8960b26](https://github.com/jarv/newsgoat/commit/8960b265dbd0518357df701056bea21a16619804))

# [1.7.0](https://github.com/jarv/newsgoat/compare/v1.6.4...v1.7.0) (2025-10-06)


### Bug Fixes

* Fixes log view attribute display ([87737a2](https://github.com/jarv/newsgoat/commit/87737a296ccd1a2f183ae62362888b34bc9abf06))
* Fixes status bar positioning on log messages ([f674d01](https://github.com/jarv/newsgoat/commit/f674d0182b94d45a82e48c2e96f43451907ec1ad))


### Features

* Feed auto-discovery for GitLab and GitHub URLs ([fb3626a](https://github.com/jarv/newsgoat/commit/fb3626a46eef56583ad0a2c2661f783e0f8b7ed4))
* Improve log summaries in the log view ([17cd932](https://github.com/jarv/newsgoat/commit/17cd932d997d51e4087091aff0bf7784aae6a623))

## [1.6.4](https://github.com/jarv/newsgoat/compare/v1.6.3...v1.6.4) (2025-10-06)


### Bug Fixes

* **deps:** update module golang.org/x/net to v0.44.0 ([#13](https://github.com/jarv/newsgoat/issues/13)) ([c7a5f6b](https://github.com/jarv/newsgoat/commit/c7a5f6bde46dd49e72db59084472764b11175040))

## [1.6.3](https://github.com/jarv/newsgoat/compare/v1.6.2...v1.6.3) (2025-10-06)


### Bug Fixes

* **deps:** update github.com/charmbracelet/lipgloss digest to b146a47 ([#12](https://github.com/jarv/newsgoat/issues/12)) ([991a2c3](https://github.com/jarv/newsgoat/commit/991a2c31e1b05d4b0b4c8319b3ff5b60ebde6bad))

## [1.6.2](https://github.com/jarv/newsgoat/compare/v1.6.1...v1.6.2) (2025-10-04)


### Bug Fixes

* Improves link references in the article view ([21975e7](https://github.com/jarv/newsgoat/commit/21975e7926650defaf61c50ea82a8a266bcd743d))

## [1.6.1](https://github.com/jarv/newsgoat/compare/v1.6.0...v1.6.1) (2025-10-04)


### Bug Fixes

* Clear feed error when getting a 304 not-modified ([e73e4da](https://github.com/jarv/newsgoat/commit/e73e4da030eeed8094099e327df21edff7b6cbb8))

# [1.6.0](https://github.com/jarv/newsgoat/compare/v1.5.0...v1.6.0) (2025-10-03)


### Features

* Adds debug flag for debug log messages ([35f20a4](https://github.com/jarv/newsgoat/commit/35f20a4266c777de019b9095465330b853de3755))

# [1.5.0](https://github.com/jarv/newsgoat/compare/v1.4.0...v1.5.0) (2025-10-03)


### Features

* Updates defaults and improves first-run ([199bcb1](https://github.com/jarv/newsgoat/commit/199bcb1c7c123332c61026bb53eb348d7d42cb91))

# [1.4.0](https://github.com/jarv/newsgoat/compare/v1.3.0...v1.4.0) (2025-10-03)


### Features

* Changes default reload concurrency from 2 to 4 ([0ffbdc8](https://github.com/jarv/newsgoat/commit/0ffbdc8e48b1c8b69aaa609de6dbb9ecb851c371))

# [1.3.0](https://github.com/jarv/newsgoat/compare/v1.2.0...v1.3.0) (2025-10-02)


### Features

* Add URLs from the tui ([5f93ad0](https://github.com/jarv/newsgoat/commit/5f93ad05ab43d89774290bf51fe5aa3d897dc511))

# [1.2.0](https://github.com/jarv/newsgoat/compare/v1.1.2...v1.2.0) (2025-10-02)


### Features

* Adds shift-n to toggle read status on items ([e91b31e](https://github.com/jarv/newsgoat/commit/e91b31ed54e98acdb4f5c9cf64311ab8cf77af8e))
* Adds the 'add' command to add urls to the feed ([31eedc7](https://github.com/jarv/newsgoat/commit/31eedc795a3dfc995df0ac42668d592724e0cb3d))
* Adds URL reloading ([5f843ba](https://github.com/jarv/newsgoat/commit/5f843bad1a5e31614488f203dad2576bb9033baa))

## [1.1.2](https://github.com/jarv/newsgoat/compare/v1.1.1...v1.1.2) (2025-10-02)


### Bug Fixes

* **deps:** update github.com/charmbracelet/lipgloss digest to 970a4b8 ([#10](https://github.com/jarv/newsgoat/issues/10)) ([046702e](https://github.com/jarv/newsgoat/commit/046702ebc94e55f667223e70ed2436687f5ea057))

## [1.1.1](https://github.com/jarv/newsgoat/compare/v1.1.0...v1.1.1) (2025-10-01)


### Bug Fixes

* **deps:** update module github.com/charmbracelet/lipgloss to v2 ([#7](https://github.com/jarv/newsgoat/issues/7)) ([eeb86b0](https://github.com/jarv/newsgoat/commit/eeb86b05035e9bbde9ae83e4e7d294a2ee07cc79))

# [1.1.0](https://github.com/jarv/newsgoat/compare/v1.0.3...v1.1.0) (2025-10-01)


### Features

* Adds urlFile option and urls.example ([ba14c7c](https://github.com/jarv/newsgoat/commit/ba14c7c9d1ce65dc7ee51c5bf13c4682daa8d30c))

## [1.0.3](https://github.com/jarv/newsgoat/compare/v1.0.2...v1.0.3) (2025-10-01)


### Bug Fixes

* **deps:** update module github.com/johanneskaufmann/html-to-markdown to v2 ([4fdca7e](https://github.com/jarv/newsgoat/commit/4fdca7ec75537c8510d9b9c737785b31736d948b))

## [1.0.2](https://github.com/jarv/newsgoat/compare/v1.0.1...v1.0.2) (2025-10-01)


### Bug Fixes

* **deps:** update module github.com/charmbracelet/lipgloss to v2 ([#4](https://github.com/jarv/newsgoat/issues/4)) ([dd538c7](https://github.com/jarv/newsgoat/commit/dd538c793773027b8578d9578dc47dad0c35d51b))

## [1.0.1](https://github.com/jarv/newsgoat/compare/v1.0.0...v1.0.1) (2025-10-01)


### Bug Fixes

* Fix build for release ([10d1118](https://github.com/jarv/newsgoat/commit/10d1118fd244faaffaaa1c239d9ca7a792b228ab))
* Switch to github.com/ncruces/go-sqlite3 ([8979b90](https://github.com/jarv/newsgoat/commit/8979b907957ef7dc215cd98ecd3dbcc119d14edd))

# 1.0.0 (2025-10-01)


### Features

* First release ([aa991c8](https://github.com/jarv/newsgoat/commit/aa991c8041ddc9ff05e8599727e0edcbc7665e6f))
