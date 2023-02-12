#!/bin/bash

# build the application
GOOS=linux GOARCH=amd64 go build || exit 1

# deploy to GoServer
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com 'rm aws-web-server'
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com 'sudo rm webserver.log'
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com 'mkdir webserver'
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com 'mkdir templates'
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com 'mkdir assets'
scp -i "$HOME/.ssh/20230129-aws.pem" ./webserver/* ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com:webserver/
scp -i "$HOME/.ssh/20230129-aws.pem" ./templates/* ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com:templates/
scp -i "$HOME/.ssh/20230129-aws.pem" ./assets/* ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com:assets/
scp -i "$HOME/.ssh/20230129-aws.pem" ./aws-web-server ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com:
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com 'sudo chmod 777 aws-web-server'
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com 'sudo systemctl restart aws-web-server.service'

# deploy to GoSever-001
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com 'rm aws-web-server'
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com 'sudo rm webserver.log'
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com 'mkdir webserver'
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com 'mkdir templates'
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com 'mkdir assets'
scp -i "$HOME/.ssh/20230129-aws.pem" ./webserver/* ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com:webserver/
scp -i "$HOME/.ssh/20230129-aws.pem" ./templates/* ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com:templates/
scp -i "$HOME/.ssh/20230129-aws.pem" ./assets/* ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com:assets/
scp -i "$HOME/.ssh/20230129-aws.pem" ./aws-web-server ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com:
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com 'sudo chmod 777 aws-web-server'
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com 'sudo systemctl restart aws-web-server.service'