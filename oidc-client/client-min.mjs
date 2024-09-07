import { Issuer, custom } from 'openid-client';
import https from 'https';
import { readFileSync } from 'fs';
import { join } from 'path';

(async () => {

  // Set up the custom agent for HTTPS requests
  const httpsAgent = new https.Agent({
    rejectUnauthorized: false // Ignore self-signed certificates for localhost (not recommended in production)
  });

  // Set custom HTTPS options
  custom.setHttpOptionsDefaults({
    agent: httpsAgent
  });

  // Discover the issuer's metadata (assuming your authorization server is running at localhost:3000)
  const issuer = await Issuer.discover('https://localhost:3000');

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

  // Now call the /trust-score API with the access token
  const requestOptions = {
    hostname: 'localhost',
    port: 4000,
    path: '/api/trust-score',
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${tokenSet.access_token}`,
      'Content-Type': 'application/json'
    },
    agent: httpsAgent, // Use the HTTPS agent
  };

  // Define the request body (sending a mobile number to the trust-score API)
  const requestBody = JSON.stringify({
    mobile_number: '+1234567890' // Example E.164 formatted mobile number
  });

  const req = https.request(requestOptions, (res) => {
    let data = '';

    // Collect the data from the response
    res.on('data', (chunk) => {
      data += chunk;
    });

    // Log the complete response when it's done
    res.on('end', () => {
      console.log('Response from /trust-score API:', data);
    });
  });

  // Handle any errors
  req.on('error', (error) => {
    console.error('Error making the request:', error);
  });

  // Write the request body and end the request
  req.write(requestBody);
  req.end();

})();