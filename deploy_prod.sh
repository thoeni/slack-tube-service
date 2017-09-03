#!/bin/bash
ssh ec2-user@services.thoeni.io 'pkill -f slack-tube-service'
./mkbin.sh linux
scp dist/slack-tube-service-linux-amd64 ec2-user@services.thoeni.io:~/
ssh ec2-user@services.thoeni.io screen -d -m './slack-tube-service-linux-amd64 &'
