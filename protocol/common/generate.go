// Package commonpb 保存客户端和服务端共同遵守的 Protobuf 契约。
package commonpb

// 生成环境固定使用 protoc 24.0 和 protoc-gen-go 1.31.0；升级时必须提交可重复生成的差异。
//go:generate protoc --go_out=. --go_opt=paths=source_relative errorcode.proto
