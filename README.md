# OpenTelemetry Schema Utilities

[![Go Reference](https://pkg.go.dev/badge/github.com/MrAlias/otel-schema-utils.svg)](https://pkg.go.dev/github.com/MrAlias/otel-schema-utils)

This repository provides conversion utilities for [OpenTelemetry Go] types using [OpenTelemetry schemas].

:construction: This repository is a work in progress and not production ready.

## Getting Started

Start by importing the project.

```go
import "github.com/MrAlias/otel-schema-utils/schema"
```

From there, construct a new `Converter` using the default client.

```go
// Passing nil here means the Converter will use schema.DefaultClient.
conv := schema.NewConverter(nil)
```

Use the `Converter` to convert [OpenTelemetry Go] types.
For example a [`*resource.Resource`].

```go
targetURL := "https://opentelemetry.io/schemas/1.20.0"
r, err := conv.Resource(ctx, targetURL, myResource)
// Handle err and use the v1.20.0 Resource r ...
```

## Clients

`Client`s is used to fetch, cache, and parse [OpenTelemetry schemas].
These schema can be located locally or on remote hosts.
Multiple constructors are provided to accommodate this variety in schema sources.

All clients will cache schemas.
Meaning any schema they return will only be fetched once from its source.
All subsequent requests for that schema will returned the same value.

### Static client

A static client fetches a static set of schemas passed to the client when it is constructed.
These clients are created using the `NewStaticClient` function.

### Local client

A local client is a static client seeded with all OpenTelemetry schemas.

These clients are useful if you want to ensure the client does not fetch any remote schemas, but still want to support OpenTelemetry published schemas.
They are created using the `NewLocalClient` function.

### HTTP client

An HTTP client fetches schemas with an HTTP request.

These clients are useful if you publish your own schemas or want to use 3rd-party schemas other than OpenTelemetry ones.
They are created using the `NewHTTPClient` function.

### Default client

A default client is provided as the `DefaultClient` variable.
This client will use a local client for all OpenTelemetry published schema, but also use an HTTP client for all uncached schemas.

## License

This Go module is distributed under the Apache-2.0 license found in the [LICENSE](./LICENSE) file.

[OpenTelemetry Go]: https://pkg.go.dev/go.opentelemetry.io/otel
[OpenTelemetry schemas]: https://opentelemetry.io/docs/specs/otel/schemas/
[`*resource.Resource`]: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/resource#Resource
