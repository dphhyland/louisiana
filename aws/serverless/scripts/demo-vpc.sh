sam deploy \
  --template-file ../customer-vpc.yaml \
  --stack-name demo-vpc \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides Customer=demo
