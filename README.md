# Vapor

Vapor is a lightweight Go server abstraction designed to provide a common request/response API across different transports.

The goal is simple: **write application logic once and expose it through different protocols without coupling the application to a specific transport.**

> **Status: Work in Progress**

## Goals

Vapor is built around a few core goals:

- Keep the server API small and predictable.
- Separate application logic from transport details.
- Provide a common request/response model across protocols.
- Provide consistent routing and parameter handling.
- Support both synchronous request/reply and publish-only messaging.
- Make it possible to add new transports without changing application code.
- Avoid becoming an unnecessarily large or opinionated framework.

The core package provides the common abstractions used by transports:

- `Server`
- `Request`
- `Response`
- `Pattern`
- Parameters
- Headers
- Context
- Request types

## How It Works

Vapor separates the application layer from the transport layer.

At the core is the `Server` interface:

    type Server interface {
        Listen(addr string) error
        Shutdown() error
        HandleFunc(Pattern, func(Request) Response) error
        Wait()
    }

Application code registers handlers against patterns:

    server.HandleFunc("/users/:id", func(req vapor.Request) vapor.Response {
        id := req.Params()["id"]

        return vapor.NewResponse(
            200,
            vapor.WithContent([]byte("user: "+id)),
        )
    })

The handler receives a Vapor `Request` and returns a Vapor `Response`.

The handler does not need to know whether the request came from HTTP, NATS, or another transport.

Conceptually:

    Transport Request
          |
          v
    Vapor Request
          |
          v
       Handler
          |
          v
    Vapor Response
          |
          v
    Transport Response

This keeps protocol-specific logic inside the transport implementation.

## Request

A Vapor `Request` provides a common representation of an incoming request.

Depending on the transport, it can expose information such as:

- Context
- Body/content
- Subject
- Request type
- Request ID
- Method
- Parameters
- Route pattern
- Query string
- Headers

Each transport is responsible for translating its native request into this common representation.

## Response

Handlers return a Vapor `Response`.

A response contains information such as:

- Status code
- Content
- Headers
- Context

The transport converts the Vapor response into the appropriate protocol-specific response.

This means application code does not need to construct HTTP responses, NATS messages, or other protocol-specific response objects directly.

## Routing

Routes are represented using `Pattern`.

Patterns can contain named parameters using the `:` prefix:

    /users/:id

The router extracts the parameter and exposes it through the request:

    id := req.Params()["id"]

The same pattern and parameter model can be used regardless of the underlying transport.

## Transports

Vapor currently provides two transport implementations:

- HTTP
- NATS

Additional transports are planned:

- gRPC — forthcoming
- WebSockets — forthcoming

The transport layer is designed so that adding another protocol does not require changing the application-level request/response model.

### HTTP

The HTTP transport converts HTTP requests into Vapor requests.

The flow is:

    HTTP Request
         |
         v
    Vapor Request
         |
         v
      Handler
         |
         v
    Vapor Response
         |
         v
    HTTP Response

HTTP-specific details remain inside the HTTP transport.

### NATS

The NATS transport maps NATS subjects to Vapor patterns.

For example:

    users.:id

A NATS message is converted into a Vapor request, passed to the handler, and converted back into a NATS response when a reply is required.

Vapor supports both publish-only and request/reply semantics:

    TypePublishOnly
    TypeRequiresReply

### gRPC

A gRPC transport is planned and forthcoming.

The intention is to expose gRPC services through the same Vapor application-level abstractions while keeping protobuf and gRPC-specific details inside the transport.

### WebSockets

A WebSocket transport is also planned and forthcoming.

The intention is to provide the same Vapor request/response model for WebSocket messages while keeping connection and protocol-specific behavior inside the WebSocket transport.

## Architecture

Vapor intentionally keeps the core small.

It does not try to replace HTTP servers, NATS clients, gRPC implementations, or WebSocket libraries.

Instead, Vapor provides a common application layer over them.

    Application
         |
         v
    +-----------+
    |   Vapor   |
    |-----------|
    | Request   |
    | Response  |
    | Pattern   |
    | Server    |
    +-----+-----+
          |
     +----+----+---------+-------------+
     |         |         |             |
     v         v         v             v
    HTTP      NATS      gRPC       WebSockets
                        (soon)        (soon)

The application depends on Vapor's abstractions.

Transport-specific code handles communication with the underlying protocol.

## Why Vapor?

Different protocols solve different problems, but application code often still revolves around the same basic concept:

    Request -> Handler -> Response

HTTP, NATS, gRPC, and WebSockets have very different semantics and capabilities.

Vapor does not attempt to make them identical. Instead, it provides a common abstraction for the parts that can reasonably be shared.

This allows an application to remain focused on its behavior while the transport handles protocol-specific concerns.

The intended separation is:

    Application logic
          |
          v
        Vapor
          |
          v
       Transport
          |
          v
       Protocol

## Installation

    go get github.com/vedadiyan/vapor

## Example

A handler can be registered using a Vapor pattern:

    server.HandleFunc("/users/:id", func(req vapor.Request) vapor.Response {
        id := req.Params()["id"]

        return vapor.NewResponse(
            200,
            vapor.WithContent([]byte("user: "+id)),
        )
    })

The application-level handler deals with Vapor's `Request` and `Response` types rather than a transport-specific API.

As additional transports are implemented, the same application model is intended to work across them.

## Still a Work in Progress

Vapor is **not production-ready**.

This repository is an active design and implementation project. The API and architecture are expected to evolve as the underlying ideas are tested and refined.

Expect:

- Breaking API changes.
- Incomplete features.
- Missing documentation.
- Limited test coverage in some areas.
- Transport behavior that may change.
- Changes to routing semantics.
- Changes to request and response structures.
- Changes to package organization.
- Performance characteristics that have not yet been fully established.

The forthcoming gRPC and WebSocket transports are not currently available and should be considered part of the roadmap rather than stable functionality.

The current priority is getting the core abstractions and transport model right before committing to long-term API stability.

If you use Vapor at this stage, expect to read the source code and adapt to changes.

## Roadmap

Current and planned transports:

- [x] HTTP
- [x] NATS
- [ ] gRPC
- [ ] WebSockets

The roadmap is intentionally small. More transports may be added in the future if they fit the core abstraction without making Vapor unnecessarily complex.

## Contributing

Contributions, ideas, and experimentation are welcome.

Because the project is still evolving, larger architectural changes should be discussed before implementation.

The main goal is to keep Vapor:

- Small
- Simple
- Transport-independent
- Predictable
- Easy to understand

## License

License information will be added as the project matures.