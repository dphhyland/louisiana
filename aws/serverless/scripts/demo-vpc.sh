sam deploy \
  --template-file ../customer-vpc.yaml \
  --stack-name trust-score-demo \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides Customer=trust-score-demo 
