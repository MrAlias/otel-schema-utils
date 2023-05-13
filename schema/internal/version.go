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

package internal

import (
	"net/url"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func Version(schemaURL string) (*semver.Version, error) {
	// https://github.com/open-telemetry/oteps/blob/main/text/0152-telemetry-schemas.md#schema-url
	u, err := url.Parse(schemaURL)
	if err != nil {
		return nil, err
	}

	return semver.NewVersion(u.Path[strings.LastIndex(u.Path, "/")+1:])
}
