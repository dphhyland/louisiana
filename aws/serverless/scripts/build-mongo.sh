docker pull mongo:latest
docker tag mongo:latest 615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/mock-telco-mongo:latest
aws ecr get-login-password --region ap-southeast-2 | docker login --username AWS --password-stdin 615299729910.dkr.ecr.ap-southeast-2.amazonaws.com
docker push 615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/mock-telco-mongo:latest