#!/bin/bash

WORKSPACE=..
LUBAN_DLL=$WORKSPACE/Tools/Luban/Luban.dll
CONF_ROOT=.

dotnet $LUBAN_DLL \
    -t server \
	-c go-json \
    -d json \
    --conf $CONF_ROOT/luban.conf \
	-x outputCodeDir=tabledef \
    -x outputDataDir=$WORKSPACE/bin/config/dev/datas \
    -x lubanGoModule=demo/luban
	