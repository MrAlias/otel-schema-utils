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

package schema

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/schema/v1.1/ast"
)

var (
	schemaURL  = "https://opentelemetry.io/schemas/1.0.0"
	schemaYAML = `
file_format: 1.1.0
schema_url: ` + schemaURL + `
versions:
  1.0.0:
`
)

func assertGet(t *testing.T, schemaURL string) func(s *ast.Schema, err error) {
	return func(s *ast.Schema, err error) {
		t.Helper()
		if assert.NoError(t, err) {
			if assert.NotNil(t, s) {
				assert.Equal(t, schemaURL, s.SchemaURL)
			}
		}
	}
}

func TestDefault(t *testing.T) {
	msg := new(string)
	*msg = schemaYAML
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, *msg)
	}))
	t.Cleanup(ts.Close)

	client := defaultClient(ts.Client())

	ctx := context.Background()
	validate := assertGet(t, schemaURL)
	validate(client.Get(ctx, ts.URL))

	// Cache miss.
	validate(client.Get(ctx, ts.URL+"/extra"))

	// Cache hit. This will fail to parse if the HTTP request is actually made.
	*msg = "first"
	validate(client.Get(ctx, ts.URL))

	// Local cache hit. This will fail to parse if the HTTP request is actually
	// made based on msg value above.
	url := "https://opentelemetry.io/schemas/1.20.0"
	require.Contains(t, local, url)
	assertGet(t, url)(client.Get(ctx, url))
}

func TestHTTP(t *testing.T) {
	msg := new(string)
	*msg = schemaYAML
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, *msg)
	}))
	t.Cleanup(ts.Close)

	client := NewHTTPClient(ts.Client())

	ctx := context.Background()
	validate := assertGet(t, schemaURL)
	validate(client.Get(ctx, ts.URL))

	// Cache miss.
	validate(client.Get(ctx, ts.URL+"/extra"))

	// Cache hit. This will fail to parse if the HTTP request is actually made.
	*msg = "first"
	validate(client.Get(ctx, ts.URL))
}

func TestLocal(t *testing.T) {
	ctx := context.Background()
	client := NewLocalClient()

	// For every YAML file in the schemas dir, ensure there is an entry in local.
	root := "./internal/schemas"
	fileSystem := os.DirFS(root)
	err := fs.WalkDir(fileSystem, ".", func(p string, _ fs.DirEntry, err error) error {
		if err != nil || path.Ext(p) != suffix {
			return err
		}

		v := strings.TrimSuffix(path.Base(p), suffix)
		url := prefix + v

		assertGet(t, url)(client.Get(ctx, url))
		return nil
	})
	require.NoError(t, err)
}
