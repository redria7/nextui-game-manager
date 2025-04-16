#!/bin/zsh
rm ./game-manager.log || true

adb pull "/mnt/SDCARD/Tools/tg5040/Game Manager.pak/game-manager.log" ./game-manager.log

printf "All done!"

printf "\a"
