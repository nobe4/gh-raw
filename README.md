# `gh-raw` 🍖

> Get the raw HTTP query

[![Go Reference](https://pkg.go.dev/badge/github.com/nobe4/gh-raw.svg)](https://pkg.go.dev/github.com/nobe4/gh-raw)
[![CI](https://github.com/nobe4/gh-raw/actions/workflows/ci.yml/badge.svg)](https://github.com/nobe4/gh-raw/actions/workflows/ci.yml)

> [!IMPORTANT]
> Under heavy development, expect nothing.
> PRs/issues welcome!

# Examples

```shell
$ ./gh-raw api /user
curl -X GET \
  https://api.github.com/user \
  -H "Accept: */*" \
  -H "Accept-Encoding: gzip" \
  -H "Content-Type: application/json; charset=utf-8" \
  -H "Time-Zone: Europe/Berlin" \
  -H "User-Agent: GitHub CLI DEV" \
  -H "Authorization: token $GITHUB_TOKEN"
```

# Install

```shell
gh extension install nobe4/gh-raw
```
