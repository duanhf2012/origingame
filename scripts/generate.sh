#!/usr/bin/env sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

# 解析 protoc 与 protoc-gen-go：优先使用 scripts/tools 下本地准备的 Linux 工具，
# 不依赖 PATH 设置；本地工具缺失时回退到 PATH 中已安装的同名命令。
tools_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)/tools
if [ -x "$tools_dir/protoc" ]; then
    protoc_cmd="$tools_dir/protoc"
elif command -v protoc >/dev/null 2>&1; then
    protoc_cmd=protoc
else
    echo "protoc 未找到：scripts/tools/protoc 不存在，PATH 中也没有 protoc。" >&2
    echo "请将 protoc 24.0 的 Linux 版放入 scripts/tools，或将其加入 PATH 后重试。" >&2
    exit 1
fi
if [ -x "$tools_dir/protoc-gen-go" ]; then
    protoc_gen_go_cmd="$tools_dir/protoc-gen-go"
elif command -v protoc-gen-go >/dev/null 2>&1; then
    protoc_gen_go_cmd=protoc-gen-go
else
    echo "protoc-gen-go 未找到：scripts/tools/protoc-gen-go 不存在，PATH 中也没有 protoc-gen-go。" >&2
    echo "请将 protoc-gen-go v1.31.0 的 Linux 版放入 scripts/tools，或将其加入 PATH 后重试。" >&2
    exit 1
fi

protoc_version=$("$protoc_cmd" --version)
protoc_gen_go_version=$("$protoc_gen_go_cmd" --version)
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
    "$protoc_cmd" --go_out=. --go_opt=paths=source_relative ./*.proto
)

if find protocol/rpc -maxdepth 1 -type f -name '*.proto' -print -quit 2>/dev/null | grep -q .; then
    (
        cd protocol/rpc
        "$protoc_cmd" --go_out=. --go_opt=paths=source_relative ./*.proto
    )
fi

if find protocol/rpc -maxdepth 1 -type f -name '*.go' -print -quit 2>/dev/null | grep -q .; then
    go run github.com/duanhf2012/origin/v3/cmd/origingen rpc ./protocol/rpc
fi

go fmt ./protocol/...
