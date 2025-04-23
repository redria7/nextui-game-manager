#!/bin/zsh
adb push ./game-manager "/mnt/SDCARD/Tools/tg5040/Game Manager.pak"
adb push ./config.yml "/mnt/SDCARD/Tools/tg5040/Game Manager.pak"

adb shell rm "/mnt/SDCARD/Tools/tg5040/Game\ Manager.pak/game-manager.log" || true

printf "Game Manager has been pushed to device!"

printf "\a"
