sam deploy \
  --template-file ../trust-score-as.yaml \
  --stack-name trust-score-as \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides EnvironmentName=dev \
  JwksUri=https://sandbox.as.trustframeworks.io/.well-known/openid-configuration \
  DomainName=sandbox.as.trustframeworks.io \
  CertificateArn=arn:aws:acm:ap-southeast-2:615299729910:certificate/c8c989fd-1eec-4a88-913e-19bb202acda5 \
  Route53HostedZoneId=Z04859113RMVSAYT46EAR \
  LoadBalancerDNS=demo-e-Appli-7Fs5eLrJfaRb-842302753.ap-southeast-2.elb.amazonaws.com:8080
