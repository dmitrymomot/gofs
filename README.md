# gofs

[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/dmitrymomot/gofs)](https://github.com/dmitrymomot/gofs)
[![Tests](https://github.com/dmitrymomot/gofs/actions/workflows/tests.yml/badge.svg)](https://github.com/dmitrymomot/gofs/actions/workflows/tests.yml)
[![CodeQL Analysis](https://github.com/dmitrymomot/gofs/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/dmitrymomot/gofs/actions/workflows/codeql-analysis.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmitrymomot/gofs)](https://goreportcard.com/report/github.com/dmitrymomot/gofs)
[![Go Reference](https://pkg.go.dev/badge/github.com/dmitrymomot/gofs.svg)](https://pkg.go.dev/github.com/dmitrymomot/gofs)
[![License](https://img.shields.io/github/license/dmitrymomot/gofs)](https://github.com/dmitrymomot/gofs/blob/main/LICENSE)

File upload server based on [golang](https://go.dev/).

## Features

- [ ] Ability to use as a library or as a standalone server. \[WIP\]
- [x] Upload single file to S3-compatible storage.
- [x] Multi-part upload to S3-compatible storage.
- [ ] Universal upload HTTP handler, which can be used for single or multi-part upload. \[WIP\]


### TODO

- [ ] Upload multiple files to S3-compatible storage.
- [ ] Redis DB adapter for multi-part upload.
- [ ] Websocket transport for multi-part upload.

