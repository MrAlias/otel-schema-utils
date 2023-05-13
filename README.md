# OpenTelemetry Schema Utilities

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
conv := NewConverter(nil)
```

Use the `Converter` to convert [OpenTelemetry Go] types.

```go
targetURL := "https://opentelemetry.io/schemas/1.20.0"
r, err := conv.Resource(ctx, targetURL, myResource)
// Handle err and use the v1.20.0 Resource r ...
```

[OpenTelemetry Go]: https://pkg.go.dev/go.opentelemetry.io/otel
[OpenTelemetry schemas]: https://opentelemetry.io/docs/specs/otel/schemas/
