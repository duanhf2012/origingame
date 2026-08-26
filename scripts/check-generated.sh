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

temporary_root=$(mktemp -d)
trap 'rm -rf -- "$temporary_root"' EXIT HUP INT TERM

"$protoc_cmd" \
    --go_out="$temporary_root" \
    --go_opt=paths=source_relative \
    ./protocol/common/*.proto

for proto_file in ./protocol/common/*.proto; do
    generated_name=$(basename "$proto_file" .proto).pb.go
    if ! cmp -s "$temporary_root/protocol/common/$generated_name" "./protocol/common/$generated_name"; then
        echo "protocol/common/$generated_name 缺失或已过期，请运行 scripts/generate.sh" >&2
        exit 1
    fi
done

for generated_file in ./protocol/common/*.pb.go; do
    proto_file=./protocol/common/$(basename "$generated_file" .pb.go).proto
    if [ ! -f "$proto_file" ]; then
        echo "$generated_file 没有对应的 .proto 源文件" >&2
        exit 1
    fi
done

if find protocol/rpc -maxdepth 1 -type f -name '*.proto' -print -quit 2>/dev/null | grep -q .; then
    "$protoc_cmd" \
        --go_out="$temporary_root" \
        --go_opt=paths=source_relative \
        ./protocol/rpc/*.proto

    for proto_file in ./protocol/rpc/*.proto; do
        generated_name=$(basename "$proto_file" .proto).pb.go
        if ! cmp -s "$temporary_root/protocol/rpc/$generated_name" "./protocol/rpc/$generated_name"; then
            echo "protocol/rpc/$generated_name 缺失或已过期，请运行 scripts/generate.sh" >&2
            exit 1
        fi
    done

    for generated_file in ./protocol/rpc/*.pb.go; do
        proto_file=./protocol/rpc/$(basename "$generated_file" .pb.go).proto
        if [ ! -f "$proto_file" ]; then
            echo "$generated_file 没有对应的 .proto 源文件" >&2
            exit 1
        fi
    done
fi

if find protocol/rpc -maxdepth 1 -type f -name '*.go' -print -quit 2>/dev/null | grep -q .; then
    go run github.com/duanhf2012/origin/v3/cmd/origingen rpc --check ./protocol/rpc
fi
