sam deploy \
  --template-file ../customer-ecs.yaml \
  --stack-name demo-ecs \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides Customer=demo VpcStack=demo-vpc
