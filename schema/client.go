// Copyright 2023 Tyler Yahn (MrAlias)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package schema // import "github.com/MrAlias/otel-schema-utils/schema"

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"

	sUtil "go.opentelemetry.io/otel/schema/v1.1"
	"go.opentelemetry.io/otel/schema/v1.1/ast"
)

//go:embed internal/schemas/*.yaml
var schemas embed.FS

const (
	prefix = "https://opentelemetry.io/schemas/"
	suffix = ".yaml"
)

var local = func() map[string]*ast.Schema {
	data := make(map[string]*ast.Schema)
	err := fs.WalkDir(schemas, ".", func(p string, _ fs.DirEntry, err error) error {
		if err != nil || path.Ext(p) != suffix {
			return err
		}

		f, err := schemas.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()

		s, err := sUtil.Parse(f)
		if err != nil {
			return err
		}

		v := strings.TrimSuffix(path.Base(p), suffix)
		url := prefix + v
		data[url] = s
		return nil
	})
	if err != nil {
		panic(err)
	}
	return data
}()

// Client is used to fetch, cache, and parse schemas.
type Client struct {
	cache

	fallback func(context.Context, string) (*ast.Schema, error)
}

func (c *Client) init() {
	if c.fallback == nil {
		c.fallback = DefaultClient.fallback
	}
}

// Get fetches and parses the OpenTelemetry schema identified by schemaURL.
func (c *Client) Get(ctx context.Context, schemaURL string) (*ast.Schema, error) {
	c.init()
	return c.lookup(schemaURL, func() (*ast.Schema, error) {
		return c.fallback(ctx, schemaURL)
	})
}

// DefaultClient is the default [Client] used to fetch schemas. All
// OpenTelemetry published schema (i.e. ones with Schema Family identifiers
// https://opentelemetry.io/schemas/<version>) are fetched from a local
// pre-built cache. All other schema are fetched using the default HTTP client
// at the remote schema URL.
var DefaultClient = defaultClient(http.DefaultClient)

func defaultClient(c *http.Client) *Client {
	httpFback := httpFallback(c)
	return &Client{
		fallback: func(ctx context.Context, url string) (*ast.Schema, error) {
			// Check static local first.
			if s, ok := local[url]; ok {
				return s, nil
			}
			return httpFback(ctx, url)
		},
	}
}

var errMissLocal = errors.New("schema not in local cache")

// NewStaticClient returns a [Client] that will only return schemas contained
// in the passed data. The passed data is a map of schema URL to schema.
func NewStaticClient(data map[string]*ast.Schema) *Client {
	return &Client{
		cache: newCache(data),
		fallback: func(_ context.Context, url string) (*ast.Schema, error) {
			return nil, fmt.Errorf("%w: %s", errMissLocal, url)
		},
	}
}

// NewLocalClient is a static [Client] pre-loaded with all OpenTelemetry
// published schema (i.e. ones with Schema Family identifiers
// https://opentelemetry.io/schemas/<version>).
func NewLocalClient() *Client { return NewStaticClient(local) }

// NewHTTPClient returns a [Client] that fetches all requested schema via an
// HTTP request to the provided schema URL. All fetched schemas are cached.
func NewHTTPClient(httpClient *http.Client) *Client {
	return &Client{fallback: httpFallback(httpClient)}
}

func httpFallback(c *http.Client) func(ctx context.Context, url string) (*ast.Schema, error) {
	if c == nil {
		c = http.DefaultClient
	}
	return func(ctx context.Context, url string) (*ast.Schema, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, err
		}
		resp, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return sUtil.Parse(resp.Body)
	}
}

type cache struct {
	data map[string]*ast.Schema
}

func newCache(data map[string]*ast.Schema) cache {
	return cache{data}
}

func (c *cache) lookup(key string, f func() (*ast.Schema, error)) (*ast.Schema, error) {
	if c.data == nil {
		s, err := f()
		if err != nil {
			return nil, errLookup{err}
		}
		c.data = map[string]*ast.Schema{
			key: s,
		}
		return s, nil
	}

	if s, ok := c.data[key]; ok {
		return s, nil
	}

	s, err := f()
	if err != nil {
		return nil, errLookup{err}
	}
	c.data[key] = s
	return s, nil
}

type errLookup struct {
	err error
}

func (e errLookup) Error() string {
	return fmt.Sprintf("schema not found: %s", e.err.Error())
}

func (e errLookup) Unwrap() error {
	return e.err
}
