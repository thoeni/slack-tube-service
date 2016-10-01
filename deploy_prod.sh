ssh ec2-user@thoeni.io 'pkill -f slack-tube-service'
./mkbin.sh
scp bin/slack-tube-service-linux-amd64 ec2-user@thoeni.io:~/
ssh ec2-user@thoeni.io screen -d -m './slack-tube-service-linux-amd64 &'
