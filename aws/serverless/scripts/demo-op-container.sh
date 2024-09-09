sam deploy \
  --template-file ../container-deployment.yaml \
  --stack-name demo-op-container \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides Customer=demo VpcStack=demo-vpc EcsClusterStack=demo-ecs Image=615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/telco-trust-score:latest ContainerName=trust-score-op
