package search

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/togo-framework/togo"
)

// sqlSearcher is the default (ParadeDB) driver. It stores documents in a
// search_documents table and queries with the dialect's case-insensitive LIKE.
// On a real ParadeDB instance the same table can be backed by a BM25 index and
// queried with the @@@ operator; the portable path keeps dev (SQLite) working.
type sqlSearcher struct {
	k     *togo.Kernel
	ready bool
}

func newSQLSearcher(k *togo.Kernel) *sqlSearcher {
	s := &sqlSearcher{k: k}
	s.ensure(context.Background())
	return s
}

func (s *sqlSearcher) ensure(ctx context.Context) {
	db, err := s.k.SQL(ctx)
	if err != nil {
		return
	}
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS search_documents (
		idx text NOT NULL,
		doc_id text NOT NULL,
		content text NOT NULL,
		body text NOT NULL,
		PRIMARY KEY (idx, doc_id)
	)`)
	s.ready = err == nil
}

func (s *sqlSearcher) Index(ctx context.Context, index, id string, doc map[string]any) error {
	db, err := s.k.SQL(ctx)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(doc)
	content := flatten(doc)
	d := s.k.Dialect()
	// Upsert (ON CONFLICT works on SQLite and Postgres).
	q := "INSERT INTO search_documents (idx, doc_id, content, body) VALUES (" + //#nosec G202 -- dialect placeholders only; values parameterized
		d.Placeholder(1) + ", " + d.Placeholder(2) + ", " + d.Placeholder(3) + ", " + d.Placeholder(4) +
		") ON CONFLICT (idx, doc_id) DO UPDATE SET content = " + d.Placeholder(3) + ", body = " + d.Placeholder(4)
	_, err = db.ExecContext(ctx, q, index, id, content, string(body))
	return err
}

func (s *sqlSearcher) Search(ctx context.Context, index, query string, limit int) ([]Hit, error) {
	db, err := s.k.SQL(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}
	d := s.k.Dialect()
	q := "SELECT doc_id, body FROM search_documents WHERE idx = " + d.Placeholder(1) + //#nosec G202 -- dialect placeholders only; values parameterized
		" AND content " + d.ILike + " " + d.Placeholder(2) + " LIMIT " + strconv.Itoa(limit)
	rows, err := db.QueryContext(ctx, q, index, "%"+strings.ToLower(query)+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hits []Hit
	for rows.Next() {
		var id, body string
		if err := rows.Scan(&id, &body); err != nil {
			return nil, err
		}
		var doc map[string]any
		_ = json.Unmarshal([]byte(body), &doc)
		hits = append(hits, Hit{ID: id, Score: 1, Doc: doc})
	}
	return hits, rows.Err()
}

func (s *sqlSearcher) Delete(ctx context.Context, index, id string) error {
	db, err := s.k.SQL(ctx)
	if err != nil {
		return err
	}
	d := s.k.Dialect()
	q := "DELETE FROM search_documents WHERE idx = " + d.Placeholder(1) + " AND doc_id = " + d.Placeholder(2) //#nosec G202 -- dialect placeholders only; values parameterized
	_, err = db.ExecContext(ctx, q, index, id)
	return err
}

// flatten lowercases and concatenates the document's scalar values for matching.
func flatten(doc map[string]any) string {
	var b strings.Builder
	for _, v := range doc {
		b.WriteString(strings.ToLower(toStr(v)))
		b.WriteByte(' ')
	}
	return b.String()
}

func toStr(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case nil:
		return ""
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}
