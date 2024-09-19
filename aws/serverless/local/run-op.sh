cd ../src/op
docker buildx build --platform linux/amd64 -t mock-telco-op .
docker stop mock-telco-op-container
docker rm mock-telco-op-container
docker run --platform linux/amd64 -d -p 3000:3000 --name mock-telco-op-container mock-telco-op