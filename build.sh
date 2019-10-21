#!/bin/bash
go build
go install
gox -osarch="linux/amd64" --output=/Users/mihuan/linux_go_bin/monitor
sshpass -p liuhanzeng scp /Users/mihuan/linux_go_bin/monitor liuhanzeng@172.16.1.44:mybin/monitor
