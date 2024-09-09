sam deploy \
  --template-file ../customer-ecs-demo.yaml \
  --stack-name demo-trust-score \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides Customer=demo \
  VpcStack=demo-vpc \
  EndpointStack=demo-internet-endpoint \
  DomainName=sandbox.as.trustframeworks.io \
  Route53HostedZoneId=Z04859113RMVSAYT46EAR \
  MTLSImage=615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/telco-trust-score-mtls:latest \
  OPImage=615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/telco-trust-score-op:latest \
  APIImage=615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/telco-trust-score-api:latest \
  MongoImage=615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/mock-telco-mongo:latest
