cd ../src/op
docker buildx build --platform linux/amd64 -t mock-telco-op .
docker tag mock-telco-op:latest 615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/telco-trust-score-op
aws ecr get-login-password --region ap-southeast-2 | docker login --username AWS --password-stdin 615299729910.dkr.ecr.ap-southeast-2.amazonaws.com
docker push 615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/telco-trust-score-op:latest