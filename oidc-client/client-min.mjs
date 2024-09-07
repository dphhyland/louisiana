import { Issuer, custom } from 'openid-client';
import https from 'https';
import { readFileSync } from 'fs';
import { join } from 'path';

(async () => {

  // Define MTLS credentials (client certificate and key)
  const mtlsClient = new https.Agent({
    cert: readFileSync(join(process.cwd(), '../certificates/client.crt')),  // Path to your client certificate
    key: readFileSync(join(process.cwd(), '../certificates/client.key')),    // Path to your client key
    ca: readFileSync(join(process.cwd(), '../certificates/ca.crt'))         // Path to your CA certificate
  });

  // Set up the custom agent for mutual TLS
  custom.setHttpOptionsDefaults({
    agent: mtlsClient
  });

  // Discover the issuer's metadata
  const issuer = await Issuer.discover('https://localhost:3000' );

  // Generate a client instance for private_key_jwt
  const client = new issuer.Client({
    client_id: 'localhost',
    token_endpoint_auth_method: 'private_key_jwt',
    token_endpoint_auth_signing_alg: 'RS256',
    redirect_uris: ['https://your-callback-url.com/cb'],
    response_types: ['code']
  }, {
    keys: [JSON.parse(readFileSync(join(process.cwd(), '../certificates/private-key.json'), 'utf8'))] // Private key in JWK format
  });

  // Define client credentials token request
  const tokenSet = await client.grant({
    grant_type: 'client_credentials',
    scope: 'your-scopes' // e.g. 'openid profile email'
  });

  console.log('Access Token:', tokenSet.access_token);

})();