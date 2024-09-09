sam deploy \
  --template-file ../customer-internet-endpoint.yaml \
  --stack-name demo-internet-endpoint \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides Customer=demo VpcStack=demo-vpc \
  DomainName=sandbox.as.trustframeworks.io \
  Route53HostedZoneId=Z04859113RMVSAYT46EAR 