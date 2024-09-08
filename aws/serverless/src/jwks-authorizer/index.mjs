const jwt = require('jsonwebtoken');
const jwksClient = require('jwks-rsa');

const client = jwksClient({
  jwksUri: process.env.JWKS_URI
});

function getKey(header, callback) {
  client.getSigningKey(header.kid, function(err, key) {
    const signingKey = key.publicKey || key.rsaPublicKey;
    callback(null, signingKey);
  });
}

exports.handler = async (event, context) => {
  const token = event.authorizationToken;
  
  try {
    const decoded = await new Promise((resolve, reject) => {
      jwt.verify(token, getKey, {}, function(err, decoded) {
        if (err) {
          reject(err);
        } else {
          resolve(decoded);
        }
      });
    });

    return generatePolicy('user', 'Allow', event.methodArn);
  } catch (error) {
    return generatePolicy('user', 'Deny', event.methodArn);
  }
};

function generatePolicy(principalId, effect, resource) {
  const authResponse = {
    principalId: principalId
  };
  
  if (effect && resource) {
    const policyDocument = {
      Version: '2012-10-17',
      Statement: [{
        Action: 'execute-api:Invoke',
        Effect: effect,
        Resource: resource
      }]
    };
    authResponse.policyDocument = policyDocument;
  }
  
  return authResponse;
}