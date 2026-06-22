# search

Full-text search for [togo](https://github.com/togo-framework/togo). Default driver
**ParadeDB** (Postgres BM25; portable SQL ILIKE fallback so it also runs on SQLite for
dev). Elasticsearch / OpenSearch ship as driver plugins.

```bash
togo install togo-framework/search
```

Env: SEARCH_DRIVER (paradedb|elasticsearch|opensearch).
