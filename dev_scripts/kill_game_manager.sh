#!/bin/zsh
printf "Shit broke! Killing Game Manager."
sshpass -p 'tina' ssh root@192.168.1.16 "kill  \$(pidof dlv)" > /dev/null 2>&1
sshpass -p 'tina' ssh root@192.168.1.16 "kill  \$(pidof game-manager)" > /dev/null 2>&1
