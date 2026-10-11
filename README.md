# zqnt-api-go

The public Go API of the Zequent platform: the protobuf/gRPC types customers build against, and
the telemetry mapping between the 2.x and 3.x contracts. The Go Edge SDK
([`zqnt-edge-sdk-go`](https://github.com/Zequent/zqnt-edge-sdk-go)) and Client SDK
([`zqnt-client-sdk-go`](https://github.com/Zequent/zqnt-client-sdk-go)) depend on this module only.

```bash
go get github.com/zequent/zqnt-api-go@latest
```

## Packages

| Import path | Contract |
|---|---|
| `gen/zqnt/capability/v3` | `zqnt.capability.v3`: Capability, CommandSafety, Command, CommandResult, CommandEvent |
| `gen/zqnt/common/v3` | `zqnt.common.v3`: AssetRef, RequestContext, Error, paging, GeoPoint |
| `gen/zqnt/control/v3` | `zqnt.control.v3`: `RemoteControlService` for clients |
| `gen/zqnt/edge/v3` | `zqnt.edge.v3`: `EdgeAdapterService` (adapter side), `EdgeGatewayService` (platform side) |
| `gen/zqnt/telemetry/v3` | `zqnt.telemetry.v3`: TelemetrySample and the telemetry service |
| `gen/common/...`, `gen/edge/sdk/proto`, `gen/livedata/proto`, `gen/remotecontrol/proto`, `gen/connector/proto`, `gen/missionautonomy/...`, `gen/execution/...`, `gen/devicecontrol/...`, `gen/events/proto`, `gen/media/proto` | the frozen 2.x contracts (package `zqnt`), wire-compatible with 2.0.0 |
| `telemetry` | v2 `Telemetry` ⇄ v3 `TelemetrySample`, and the catalog of telemetry detail keys |

All import paths are below `github.com/zequent/zqnt-api-go/`. The `.proto` files are in
[`proto/`](proto): `proto/v2` and `proto/v3` are the two proto roots (`-I proto/v3` for
`import "zqnt/common/v3/common.proto"`).

## Where it comes from

`proto/` is a copy of the public contracts of the platform's contract repository at the commit in
[`proto/ZQNT_PROTOS_COMMIT`](proto/ZQNT_PROTOS_COMMIT); `gen/` is generated from it by
`scripts/gen_protos.sh` with protoc 3.21.12, protoc-gen-go v1.36.11 and protoc-gen-go-grpc v1.6.1.
Both are updated by an automated PR after every contract change; do not edit them by hand. CI
regenerates `gen/` on every PR and fails if it differs.

## Versions

Releases are `vX.Y.Z` tags. Every PR build is also tagged `pr-<n>-<sha>` and can be fetched with
`go get github.com/zequent/zqnt-api-go@pr-<n>-<sha>` for review.

## License

Licensed under the [Apache License, Version 2.0](LICENSE); see [NOTICE](NOTICE).
