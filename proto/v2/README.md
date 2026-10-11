# zqnt-protos — v2 (frozen)

> **Frozen 2.x contracts.** These files are byte-identical to the `2.0.0` release (package `zqnt`,
> imports relative to the v2 root). Only bug fixes, never new features; every change must pass the
> wire + JSON breaking check against `2.0.0` in CI. New work goes into [`../v3`](../v3) —
> see the [root README](../../README.md). File references below are relative to `v2/`.

Shared Protobuf definitions for the ZQNT framework. All services (Edge SDK, Edge Adapter, RemoteControl, LiveData, Connector, MissionAutonomy) use these proto files as the single source of truth for their API contracts.

---

## Architecture Overview

```
Client (SDK)
    │
    ▼
RemoteControlService          ← client-facing command gateway
    │  routes by asset SN
    ▼
Edge Adapter                  ← dispatches to correct Edge SDK instance
    │
    ▼
Edge SDK (per vendor/device)  ← implements EdgeAdapterService
    │
    ▼
Physical Device (DJI, Autel, ROS, ...)
```

**LiveDataService** runs in parallel — Edge SDKs push telemetry upstream, clients subscribe to streams.

**ConnectorService** handles persistent storage of assets, missions, tasks and schedulers.

**MissionAutonomyService** manages mission planning and task lifecycle.

---

## Proto Files

| File | Service | Purpose |
|---|---|---|
| `common.proto` | — | Compatibility facade that public-imports the shared domain proto files |
| `base.proto` | — | Shared request/response metadata, errors and command progress |
| `asset.proto` | — | Asset/sub-asset enums and DTOs |
| `device-control-contracts.proto` | — | Shared device-control requests, responses, capabilities and live-stream contracts |
| `detection.proto` | — | Detection stream/control payloads and detection result types |
| `live-data-types.proto` | — | Live-data requests, responses, telemetry models and detection response wrappers |
| `mission-autonomy-types.proto` | — | Mission/task/scheduler enums, waypoints and typed task configs |
| `mission-autonomy-dto.proto` | — | Mission, task and scheduler DTOs |
| `mission-autonomy-contracts.proto` | — | Shared mission-autonomy requests and responses for missions, tasks and schedulers |
| `edge.proto` | `EdgeAdapterService` | Device control interface implemented by each Edge SDK; uses shared command contracts |
| `remote-control.proto` | `RemoteControlService` | Client-facing gateway that routes commands to the correct Edge Adapter; uses shared command contracts |
| `live-data.proto` | `LiveDataService` | RPC-only contract for telemetry, detections, notifications and live-stream control |
| `connector.proto` | `ConnectorService` | CRUD for assets, missions, tasks, schedulers and telemetry storage |
| `mission-autonomy.proto` | `MissionAutonomyService` | Mission and task lifecycle management |

---

## Key Conventions

### Source of Truth Rules

The proto files are the source of truth. To keep them stable and fast to evolve:

- Add reusable payloads to the matching shared domain file first. Keep `common.proto` as a public-import facade for compatibility.
- Add service-specific request wrappers only when the service really needs a different wire contract.
- Do not create `EdgeXxxRequest` and `RemoteControlXxxRequest` with the same fields for new commands. Use one of the shared `*CommandRequest` messages instead.
- Avoid compatibility wrappers while the SDKs are still early and controlled internally. Introduce versioned compatibility APIs only when external consumers need a stable legacy contract.
- Keep all proto files in the same explicit proto package (`package zqnt;`) unless there is a deliberate versioned namespace split.
- Never change field numbers or field types. Add new fields with new numbers, and reserve removed numbers/names.
- Prefer `optional` for fields where presence matters. For required business validation, validate in service code; proto3 itself does not enforce required fields.
- Enum value `0` should mean `UNSPECIFIED` for new enums. Existing enums that use a real value at `0` are kept for compatibility.
- Keep cross-domain infrastructure types in `com.zqnt.utils.common.proto`; domain-owned files use their service namespace, such as `com.zqnt.utils.mission.proto` for mission autonomy.

### Shared Command Types

New command-style RPCs should prefer the shared wrappers from `device-control-contracts.proto`:

```proto
message EmptyCommandRequest {
  RequestBase base = 1;
}

message CoordinateCommandRequest {
  RequestBase base = 1;
  GeoCoordinate coordinate = 2;
}

message TaskCommandRequest {
  RequestBase base = 1;
  string task_id = 2;
}

message CommandResponse {
  optional bool has_errors = 1;
  ResponseMeta meta = 2;
  oneof response {
    google.protobuf.Empty empty = 3;
    GlobalErrorMessage error = 4;
    CommandProgress progress = 5;
    LiveStreamStartResponse live_stream_start_response = 6;
  }
}
```

`EdgeAdapterService` and `RemoteControlService` use these shared command types directly. This keeps client SDKs and edge SDKs aligned and avoids duplicated generated classes, mappers, and tests.

### RequestBase

Every command request carries a `RequestBase`:

```proto
message RequestBase {
  string tid = 1;                      // Transaction ID for tracing
  string sn  = 2;                      // Asset serial number (routing key)
  google.protobuf.Timestamp timestamp = 3;
  optional string asset_id = 4;
  optional string external_id = 5;
  optional string client_id = 6;
  optional string user_id = 7;
}
```

The `sn` field is used by the Edge Adapter to route the command to the correct Edge SDK instance.

### Command Response Pattern

Command RPCs use one shared response envelope:

```proto
message CommandResponse {
  optional bool has_errors = 1;
  ResponseMeta meta = 2;
  oneof response {
    google.protobuf.Empty empty = 3;
    GlobalErrorMessage error = 4;
    CommandProgress progress = 5;
    LiveStreamStartResponse live_stream_start_response = 6;
  }
}
```

Domain services may still use service-specific response envelopes when they return service-specific DTOs.

### Extension Point for Integrators

Every command service exposes a `SendCustomCommand` RPC for integrator-defined commands without requiring proto changes:

```proto
rpc SendCustomCommand(CustomCommandRequest) returns (CustomCommandResponse);

message CustomCommandRequest {
  RequestBase            base       = 1;
  string                 command_id = 2;  // e.g. "asset.payload.sprayer.activate"
  google.protobuf.Struct params     = 3;  // JSON-compatible dynamic params
}
```

Custom commands and detected payloads are pushed as one complete snapshot to
`RemoteControlService.ReportAssetRuntime`.
Clients read the latest snapshot through `RemoteControlService.GetCapabilities`. The
`Capability.command_id` field uses the same canonical ID as `CustomCommandRequest.command_id`.

Capability IDs use lowercase dot-separated namespaces, for example `asset.navigation.goto`,
`asset.navigation.return_home`, or `asset.payload.gimbal.look_at`. Edge adapters are the
authoritative source and push snapshots to Remote Control on a schedule. Every payload has
a stable `payload_ref`; capabilities and custom commands address it through the same
`CapabilityTarget`. Remote Control caches snapshots with a TTL and publishes an event when
revision or freshness changes.

Unknown vendor capabilities are self-describing: `input_schema` and `output_schema` use
JSON Schema objects so client SDKs can render inputs, validate parameters, and display
results without a release for every new adapter. Stable platform commands continue to use
typed RPCs; vendor extensions are invoked through `SendCustomCommand` with the advertised
`command_id`. `Capability` is the canonical public command definition; payload inventory
DTOs describe payload identity and state and must not duplicate command or schema definitions.

Clients use `GetAssetRuntime` to load detected payloads and capabilities as one atomic
snapshot. A discovered capability can be scheduled through `DynamicCommandTaskConfigProto`;
discovery, task planning, and `SendCustomCommand` share the same `command_id`, target, and
parameter object. Adapters change `schema_version` whenever an input or output contract changes
incompatibly. Long-lived tasks copy it to `expected_schema_version` and fail validation when the
currently advertised version differs.

---

## Naming Conventions

- **Field names**: `snake_case`
- **Message names**: `PascalCase`. Use functional shared names for reusable contracts (e.g. `CoordinateCommandRequest`) and service-prefixed names only for service-specific contracts (e.g. `ConnectorResponse`)
- **Enum types**: `PascalCase` with descriptive suffix (e.g. `AssetTypeEnum`, `LiveStreamTypeEnum`)
- **Enum values**: `SCREAMING_SNAKE_CASE` prefixed with the enum type (e.g. `ASSET_TYPE_DOCK`, `LIVE_DATA_TYPE_ASSET_TELEMETRY`)
- **RPC methods**: `PascalCase` verb-first (e.g. `TakeOff`, `GetCapabilities`, `SendCustomCommand`)

For new shared messages, use a functional name instead of a service prefix, for example `TaskCommandRequest`. Service prefixes are reserved for service-specific payloads and responses.

### Remaining Technical Debt

The shared contracts are now split by domain. Remaining cleanup should focus on API stability:

- Several legacy enums use a real value at `0`; new enums should use `*_UNSPECIFIED = 0`.
- Some service responses outside the command layer still have service-specific response envelopes.
- `common.proto` should stay as a compatibility facade until all consumers import the domain files directly.

Recommended cleanup path:

1. Keep command services on shared `*CommandRequest` and `CommandResponse`.
2. Add new SDK commands only once in `device-control-contracts.proto`, then expose them from the relevant service.
3. Preserve field numbers and reserve removed fields in all future changes.
4. Introduce versioned services only after the first external SDK contract is published.

---

## Regenerating Code

Stubs are generated from these proto files for each language SDK. After changing any `.proto` file, regenerate all language stubs and update downstream services in a coordinated release.

Breaking changes (message renames, field type changes, field number changes) require a versioned release across all services simultaneously.

---

## Versioning

This repo is the single source of truth for the wire contract every language SDK (Java `zqnt-utils`,
`edge-go-sdk`/`client-go-sdk`, `edge-python-sdk`/`client-python-sdk`/`zqnt-utils-python`) generates
from — but each of those consumers also has its own, independent package version number for its own
non-proto changes. Those two numbers are **not** the same thing, and conflating them is exactly what
prompted this section: `zqnt-utils` (Java) reached `v1.3.0` and `zqnt-utils-python` reached `v2.0.0`
around the same time, encoding the same underlying contract, purely by each project's own unrelated
semver history — not because the contract itself had two different states.

**`PROTOCOL_VERSION`** (this directory) holds the current contract version as a single line — the one
number every language's generated output should be checked against, independent of that language's
own package version. Bump it whenever a message/RPC changes in a way downstream consumers need to
know about (new field, renamed message, new RPC, changed field number/type); a purely additive,
non-breaking addition can land without a bump if nothing downstream needs to react to it.

**Git tags on this repo** pin the exact commit a given `PROTOCOL_VERSION` was cut at, so "what did the
contract look like at version X" is one `git checkout <tag>` away instead of correlating submodule
pointers and commit messages by hand:

- `1.3.0` → `0e072f8` — the state `zqnt-utils` (Java) pointed to when *it* was tagged `v1.3.0`.
- `2.0.0` → the commit actually in use as of 2026-09-03 — matches `zqnt-utils` HEAD and what
  `zqnt-utils-python`'s `v2.0.0` was regenerated from. Named to match the already-established
  `zqnt-utils-python` version rather than invent a fourth numbering scheme.

When cutting a new contract version: bump `PROTOCOL_VERSION`, commit, tag the commit here with that
same version, then have each downstream SDK's generation step record which tag it built against (a
comment at the top of the generated output, or a small manifest) — see each SDK's own README for
where that's recorded.
