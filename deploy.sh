#!/bin/bash

# build and push the image to the registry
aws ecr get-login-password --region eu-north-1 | docker login --username AWS --password-stdin 906350741214.dkr.ecr.eu-north-1.amazonaws.com
docker build -t go-web-server .
docker tag go-web-server:latest 906350741214.dkr.ecr.eu-north-1.amazonaws.com/go-web-server:latest
docker push 906350741214.dkr.ecr.eu-north-1.amazonaws.com/go-web-server:latest

# deploy to GoServer
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com "aws ecr get-login-password --region eu-north-1 | docker login --username AWS --password-stdin 906350741214.dkr.ecr.eu-north-1.amazonaws.com"
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com "docker pull 906350741214.dkr.ecr.eu-north-1.amazonaws.com/go-web-server:latest"
scp -i "$HOME/.ssh/20230129-aws.pem" ./docker-compose.yml ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com:
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-13-53-40-61.eu-north-1.compute.amazonaws.com "docker compose up -d"

# deploy to GoSever-001
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com "aws ecr get-login-password --region eu-north-1 | docker login --username AWS --password-stdin 906350741214.dkr.ecr.eu-north-1.amazonaws.com"
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com "docker pull 906350741214.dkr.ecr.eu-north-1.amazonaws.com/go-web-server:latest"
scp -i "$HOME/.ssh/20230129-aws.pem" ./docker-compose.yml ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com:
ssh -i "$HOME/.ssh/20230129-aws.pem" ubuntu@ec2-16-170-236-95.eu-north-1.compute.amazonaws.com "docker compose up -d"
