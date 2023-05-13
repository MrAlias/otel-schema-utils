// Copyright 2022 Tyler Yahn (MrAlias)
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
	"errors"

	"github.com/MrAlias/otel-schema-utils/schema/internal/cmp"
	"github.com/MrAlias/otel-schema-utils/schema/internal/resconv"

	"go.opentelemetry.io/otel/sdk/resource"
)

// Converter converts OpenTelemetry objects based on OpenTelemetry schema
// translations.
type Converter struct {
	client *Client
}

// NewConverter returns a new [Converter] that uses the [Client] c to handle
// all schema fetching. If c is nil, the default client will be used.
func NewConverter(c *Client) Converter {
	if c == nil {
		c = DefaultClient
	}
	return Converter{client: c}
}

// Resource returns a copy of orig with the schema URL set to url and all
// attributes transformed based on the associated schema. If the schema
// transformation fails, or url is empty, an error is returned.
func (c Converter) Resource(ctx context.Context, schemaURL string, orig *resource.Resource) (*resource.Resource, error) {
	if schemaURL == "" {
		return nil, errors.New(`invalid schema url: ""`)
	}

	if orig == nil || orig.Len() == 0 {
		return resource.NewWithAttributes(schemaURL), nil
	}

	if orig.SchemaURL() == schemaURL {
		// Resources are immutable, just return the ptr to the same value.
		return orig, nil
	}

	comp, err := cmp.Versions(orig.SchemaURL(), schemaURL)
	if err != nil {
		return nil, err
	}
	switch comp {
	case cmp.EqualTo:
		// Resources are immutable, just return the ptr to the same value.
		return orig, nil
	case cmp.LessThan:
		s, err := c.client.Get(ctx, schemaURL)
		if err != nil {
			return nil, err
		}
		attrs := orig.Attributes()
		err = resconv.Upgrade(s, attrs)
		if err != nil {
			return nil, err
		}
		return resource.NewWithAttributes(schemaURL, attrs...), nil
	case cmp.GreaterThan:
		s, err := c.client.Get(ctx, orig.SchemaURL())
		if err != nil {
			return nil, err
		}
		attrs := orig.Attributes()
		err = resconv.Downgrade(s, schemaURL, attrs)
		if err != nil {
			return nil, err
		}
		return resource.NewWithAttributes(schemaURL, attrs...), nil
	default:
		panic("unknown schema URL comparison")
	}
}

// MergeResources creates a new resource by combining resources at the target
// schemaURL version.
//
// If there are common keys between resources the latter resource will
// overwrite the former.
//
// Any of the resources not already at schemaURL version will be appropriately
// upgraded or downgraded to match the version. An error is returned if this is
// not possible.
func (c Converter) MergeResources(ctx context.Context, schemaURL string, resources ...*resource.Resource) (*resource.Resource, error) {
	merged := resource.NewWithAttributes(schemaURL)
	for _, r := range resources {
		versioned, err := c.Resource(ctx, schemaURL, r)
		if err != nil {
			return nil, err
		}
		merged, err = resource.Merge(merged, versioned)
		if err != nil {
			return nil, err
		}
	}
	return merged, nil
}
