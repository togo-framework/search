// Package search is togo's full-text search subsystem. The default driver is
// ParadeDB (Postgres BM25; it degrades to a portable SQL ILIKE search so dev on
// SQLite works too). Elasticsearch, OpenSearch, etc. ship as driver plugins that
// call search.RegisterDriver and depend on this package.
//
// Install: `togo install togo-framework/search`.
package search

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/togo-framework/togo"
)

// Hit is a single search result.
type Hit struct {
	ID    string         `json:"id"`
	Score float64        `json:"score"`
	Doc   map[string]any `json:"doc"`
}

// Searcher indexes and queries documents grouped by index name.
type Searcher interface {
	Index(ctx context.Context, index, id string, doc map[string]any) error
	Search(ctx context.Context, index, query string, limit int) ([]Hit, error)
	Delete(ctx context.Context, index, id string) error
}

// DriverFactory builds a Searcher from the kernel.
type DriverFactory func(k *togo.Kernel) (Searcher, error)

var (
	regMu   sync.RWMutex
	drivers = map[string]DriverFactory{}
)

// RegisterDriver registers a search driver by name (call from a plugin init()).
func RegisterDriver(name string, f DriverFactory) {
	regMu.Lock()
	drivers[name] = f
	regMu.Unlock()
}

func init() {
	RegisterDriver("paradedb", func(k *togo.Kernel) (Searcher, error) { return newSQLSearcher(k), nil })
	RegisterDriver("sql", func(k *togo.Kernel) (Searcher, error) { return newSQLSearcher(k), nil })

	togo.RegisterProviderFunc("search", togo.PriorityLate, func(k *togo.Kernel) error {
		name := os.Getenv("SEARCH_DRIVER")
		if name == "" {
			name = "paradedb"
		}
		regMu.RLock()
		f, ok := drivers[name]
		regMu.RUnlock()
		if !ok {
			return fmt.Errorf("search: unknown driver %q (install its plugin?)", name)
		}
		s, err := f(k)
		if err != nil {
			return err
		}
		k.Set("search", &Service{searcher: s, driver: name})
		return nil
	})
}

// Service is the search runtime stored on the kernel (k.Get("search")).
type Service struct {
	searcher Searcher
	driver   string
}

func (s *Service) Index(ctx context.Context, index, id string, doc map[string]any) error {
	return s.searcher.Index(ctx, index, id, doc)
}
func (s *Service) Search(ctx context.Context, index, query string, limit int) ([]Hit, error) {
	return s.searcher.Search(ctx, index, query, limit)
}
func (s *Service) Delete(ctx context.Context, index, id string) error {
	return s.searcher.Delete(ctx, index, id)
}

// Driver returns the active driver name.
func (s *Service) Driver() string { return s.driver }

// FromKernel fetches the search service from the kernel container.
func FromKernel(k *togo.Kernel) (*Service, bool) {
	v, ok := k.Get("search")
	if !ok {
		return nil, false
	}
	s, ok := v.(*Service)
	return s, ok
}
