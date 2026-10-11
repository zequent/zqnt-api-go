#!/usr/bin/env bash
# Generates gen/ from the public contracts in proto/ (a copy of zqnt-protos public/, pinned by
# proto/ZQNT_PROTOS_COMMIT). The zqnt-protos generator workflow refreshes proto/ and runs this
# script; run it by hand only to check that gen/ matches proto/.
#
# Tool versions are part of the output: protoc 3.21.12 (release v21.12), protoc-gen-go v1.36.11,
# protoc-gen-go-grpc v1.6.1.
#
#   proto/v2  frozen 2.x contracts, package `zqnt`; each file's go_package ("gen/common/proto")
#             is placed under this module.
#   proto/v3  zqnt.<domain>.v3 packages with full go_package paths -> gen/zqnt/<domain>/v3.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODULE="github.com/zequent/zqnt-api-go"
V2_DIR="$ROOT/proto/v2"
V3_DIR="$ROOT/proto/v3"

expect_version() {
  local actual
  actual="$("$@" 2>&1 | awk '{print $NF}')"
  if [ "$actual" != "$EXPECTED" ]; then
    echo "$1 is $actual, expected $EXPECTED (generated code would differ)" >&2
    exit 1
  fi
}
EXPECTED=3.21.12 expect_version protoc --version
EXPECTED=v1.36.11 expect_version protoc-gen-go --version
EXPECTED=1.6.1 expect_version protoc-gen-go-grpc --version

v2_mappings=()
for f in "$V2_DIR"/*.proto; do
  pkg="$(sed -nE 's/^option go_package = "([^";]+)(;[^"]*)?";/\1/p' "$f")"
  v2_mappings+=("M$(basename "$f")=$MODULE/$pkg")
done

rm -rf "$ROOT/gen"

go_args=(--go_opt=paths=import "--go_opt=module=$MODULE")
grpc_args=(--go-grpc_opt=paths=import "--go-grpc_opt=module=$MODULE" --go-grpc_opt=require_unimplemented_servers=true)
for m in "${v2_mappings[@]}"; do
  go_args+=("--go_opt=$m")
  grpc_args+=("--go-grpc_opt=$m")
done

(cd "$V2_DIR" && protoc --proto_path=. \
  --go_out="$ROOT" "${go_args[@]}" \
  --go-grpc_out="$ROOT" "${grpc_args[@]}" \
  *.proto)

mapfile -t v3_files < <(cd "$V3_DIR" && find zqnt -name '*.proto' | sort)
(cd "$V3_DIR" && protoc --proto_path=. \
  --go_out="$ROOT" --go_opt=paths=import "--go_opt=module=$MODULE" \
  --go-grpc_out="$ROOT" --go-grpc_opt=paths=import "--go-grpc_opt=module=$MODULE" \
  --go-grpc_opt=require_unimplemented_servers=true \
  "${v3_files[@]}")

echo "Generated gen/ from zqnt-protos $(cat "$ROOT/proto/ZQNT_PROTOS_COMMIT" 2>/dev/null || echo '(unknown commit)')."
