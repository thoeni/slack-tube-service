#!/bin/bash
ssh ec2-user@services.thoeni.io 'rm slack-tube-service-linux-amd64 && wget $(wget -qO- https://api.github.com/repos/thoeni/slack-tube-service/releases/latest | grep -o "http.*linux-amd64") && sudo chmod +x slack-tube-service-linux-amd64'
ssh ec2-user@services.thoeni.io 'pkill -f slack-tube-service'
ssh ec2-user@services.thoeni.io screen -d -m './slack-tube-service-linux-amd64 &'