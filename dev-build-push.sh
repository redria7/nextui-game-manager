#!/bin/zsh
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC="zig cc -target aarch64-linux" CXX="zig c++ -target aarch64-linux" go build -o game-manager --tags extended .

adb push ./game-manager "/mnt/SDCARD/Tools/tg5040/Game Manager.pak"

echo '\a'

echo "Donzo Washington!"
