#!/usr/bin/env sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

protoc_version=$(protoc --version)
protoc_gen_go_version=$(protoc-gen-go --version)
if [ "$protoc_version" != "libprotoc 24.0" ]; then
    echo "protoc 版本必须为 24.0，当前为 $protoc_version" >&2
    exit 1
fi
if [ "$protoc_gen_go_version" != "protoc-gen-go v1.31.0" ]; then
    echo "protoc-gen-go 版本必须为 1.31.0，当前为 $protoc_gen_go_version" >&2
    exit 1
fi

(
    cd protocol/common
    protoc --go_out=. --go_opt=paths=source_relative ./*.proto
)

if find protocol/rpc -maxdepth 1 -type f -name '*.proto' -print -quit 2>/dev/null | grep -q .; then
    (
        cd protocol/rpc
        protoc --go_out=. --go_opt=paths=source_relative ./*.proto
    )
fi

if find protocol/rpc -maxdepth 1 -type f -name '*.go' -print -quit 2>/dev/null | grep -q .; then
    go run github.com/duanhf2012/origin/v3/cmd/origingen rpc ./protocol/rpc
fi

go fmt ./protocol/...
