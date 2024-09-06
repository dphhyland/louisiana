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

// Load the JWKS from your file
const jwks = JSON.parse(fs.readFileSync(path.join(__dirname, '../certificates/client.jwks.json'), 'utf8'));

// Load client certificate, key, and CA certificate
const clientCert = fs.readFileSync(path.join(__dirname, '../certificates/client.crt'));
const clientKey = fs.readFileSync(path.join(__dirname, '../certificates/client.key'));
const caCert = fs.readFileSync(path.join(__dirname, '../certificates/ca.crt'));

// Create an HTTPS agent for mTLS
const httpsAgent = new https.Agent({
  cert: clientCert, // Provide the client certificate
  key: clientKey,   // Provide the client private key
  ca: caCert,       // Optionally provide the CA certificate if needed to verify the server
  rejectUnauthorized: false // Set to true in production to validate the server's certificate
});

async function discoverAndAuthenticate() {
  try {
    const issuerUrl = `https://localhost:${process.env.PORT || 3000}`;

    // Discover the issuer configuration from the .well-known endpoint
    const issuer = await Issuer.discover(issuerUrl);
    console.log('Discovered issuer:', issuer.issuer);

    const client = new issuer.Client({
      client_id: 'localhost',
      token_endpoint_auth_method: 'self_signed_tls_client_auth', // Use private_key_jwt for client assertions
      token_endpoint_auth_signing_alg: 'RS256', 
      tls_client_auth_subject_dn: 'C=AU, ST=NSW, L=Sydney, O=MY DIGITAL ID PTY LTD, CN=localhost'
    }, { jwks });

    // Customize HTTP options for mTLS on a per-request basis
    client[custom.http_options] = (url, options) => {
      return {
        cert: clientCert,
        key: clientKey,
        ca: caCert, // Use the custom CA if needed
        rejectUnauthorized: false // Change to true in production environments
      };
    };

    // Perform the client credentials grant to obtain an access token
    const tokenSet = await client.grant({
      grant_type: 'client_credentials',
      scope: 'api:access',  // Adjust based on your server configuration
    });

    console.log('Access Token:', tokenSet.access_token);
  } catch (error) {
    console.error('Error during authentication:', error);
  }
}

discoverAndAuthenticate().catch(error => {
  console.error('Unhandled error:', error);
});