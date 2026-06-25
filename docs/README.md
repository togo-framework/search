# search — documentation

  <img src=".github/assets/togo-mark.svg" alt="togo" height="64" />

## Overview

Package search is togo's full-text search subsystem. The default driver is
ParadeDB (Postgres BM25; it degrades to a portable SQL ILIKE search so dev on
SQLite works too). Elasticsearch, OpenSearch, etc. ship as driver plugins that
call search.RegisterDriver and depend on this package.

Install: `togo install togo-framework/search`.

## Install

```bash
togo install togo-framework/search
```

Set `SEARCH_DRIVER=<provider>` and install a driver (search-algolia, …).

## Configuration

Environment variables read by this plugin (extracted from the source):

| Env var | Notes |
|---|---|
| `G` | _see provider docs_ |
| `SEARCH_DRIVER` | _see provider docs_ |

## Usage

```go
s := k.Search
s.Index(ctx, "posts", doc)
hits, _ := s.Search(ctx, "posts", "query")
```

## Links

- Marketplace: https://to-go.dev/marketplace
- Source: https://github.com/togo-framework/search
- README: ../README.md
