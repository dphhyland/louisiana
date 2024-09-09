sam deploy \
  --template-file ../container-deployment.yaml \
  --stack-name demo-op-container \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides Customer=demo \
  VpcStack=demo-vpc \
  EcsClusterStack=demo-ecs \
  MTLSImage=615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/telco-trust-score-mtls:latest \
  OPImage=615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/telco-trust-score-op:latest \
  APIImage=615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/telco-trust-score-api:latest \
  MongoImage=615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/mock-telco-mongo:latest
