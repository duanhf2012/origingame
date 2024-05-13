protoc.exe --go_out=./rpc ./rpcproto/*.proto
protoc.exe --go_out=./msg ./msgproto/*.proto
protoc.exe --go_out=../db -I ../db/ ../db/*.proto

PAUSE