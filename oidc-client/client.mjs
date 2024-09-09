import https from 'https';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { Issuer, custom } from 'openid-client';
import dotenv from 'dotenv';

dotenv.config();

// Correctly set up __filename and __dirname for ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Load client certificate, key, and CA certificate
const clientCert = fs.readFileSync(path.join(__dirname, './certificates/client.crt'));
const clientKey = fs.readFileSync(path.join(__dirname, './certificates/client.key'));
const caCert = fs.readFileSync(path.join(__dirname, './certificates/ca.crt'));

// Load the private key JWK from your file
const jwks = JSON.parse(fs.readFileSync(path.join(__dirname, './certificates/client.jwks.json'), 'utf8'));


const httpsAgent = new https.Agent({
  cert: clientCert,
  key: clientKey,
  rejectUnauthorized: true,  
});

async function discoverAndAuthenticate() {
  try {
    console.log('Discovering issuer...');
    console.log('Issuer metadata:', jwks);
    const issuerUrl = `https://sandbox.as.trustframeworks.io`;

    // Discover the issuer configuration from the .well-known endpoint
    const issuer = await Issuer.discover(issuerUrl, {
      agent: httpsAgent,
    });
    console.log('Discovered issuer:', issuer.issuer);



    // Initialize the client with private_key_jwt authentication
    const client = new issuer.Client({
      client_id: 'localhost',
      token_endpoint_auth_method: 'private_key_jwt', // Use private_key_jwt for client authentication
      token_endpoint_auth_signing_alg: 'RS256', // Define signing algorithm
      jwks, // Provide the JWKS containing private key for signing the client assertion
    });


    // Perform the client credentials grant to obtain an access token
    const tokenSet = await client.grant({
      grant_type: 'client_credentials',
      scope: 'api:access', // Adjust based on your server configuration
    });

    console.log('Access Token:', tokenSet.access_token);
  } catch (error) {
    console.error('Error during authentication:', error);
  }
}

// Start the authentication process
discoverAndAuthenticate().catch(error => {
  console.error('Unhandled error:', error);
});