# shortener

[![CI](https://github.com/ertugruldasgin/shortener/actions/workflows/ci.yml/badge.svg)](https://github.com/ertugruldasgin/shortener/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ertugruldasgin/shortener.svg)](https://pkg.go.dev/github.com/ertugruldasgin/shortener)
[![License: GPL v2](https://img.shields.io/badge/license-GPLv2-blue.svg)](LICENSE)

A self-hosted URL shortener in Go, with custom aliases, link expiry, and click
analytics. Built around a hexagonal core: the domain package knows nothing about
HTTP, PostgreSQL, or Redis.
