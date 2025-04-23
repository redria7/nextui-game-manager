#!/bin/zsh
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC="zig cc -target aarch64-linux" CXX="zig c++ -target aarch64-linux" go build -gcflags="all=-N -l"  -o game-manager --tags extended . || exit