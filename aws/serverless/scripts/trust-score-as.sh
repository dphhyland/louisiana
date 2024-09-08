sam deploy \
  --template-file ../trust-score-as.yaml \
  --stack-name trust-score-as \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides EnvironmentName=dev JwksUri=https://sandbox.as.trustframeworks.io/.well-known/openid-configuration DomainName=sandbox.as.trustframeworks.io CertificateArn=arn:aws:acm:ap-southeast-2:615299729910:certificate/c8c989fd-1eec-4a88-913e-19bb202acda5 Route53HostedZoneId=Z04859113RMVSAYT46EAR ECRRepo=615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/telco-trust-score \
  --image-repositories OPFunctionV2=615299729910.dkr.ecr.ap-southeast-2.amazonaws.com/telco-trust-score