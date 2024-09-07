// Import necessary modules and functions
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';
import { importPKCS8, exportJWK } from 'jose';  // Correct imports from jose

// Set up __filename and __dirname for ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Load client configurations from a JSON file
const clientsPath = path.join(__dirname, 'clients.json');
const clients = JSON.parse(fs.readFileSync(clientsPath, 'utf8'));

// Load the private key PEM file
const privateKeyPEM = fs.readFileSync(path.join(__dirname, '../certificates/server.key'), 'utf8');

// Asynchronously generate JWKS
async function generateJwks() {
  // Import the PEM private key as a JWK object using PKCS8 format
  const privateKey = await importPKCS8(privateKeyPEM, 'RS256');
  const privateKeyJWK = await exportJWK(privateKey);

  // Create a JWKS object containing the public part of the private key
  const jwks = { keys: [privateKeyJWK] };  // Exporting the public key in JWK format

  return jwks;
}

// Generate JWKS and configure the OIDC provider
const jwks = await generateJwks(); // Use top-level await to generate JWKS before configuration

const oidcConfiguration = {
  clients: clients,
  formats: {
    AccessToken: 'jwt', // Use JWT format for access tokens
  },
  jwks, // Automatically serves this JWKS at /.well-known/jwks.json
  features: {
    clientCredentials: { enabled: true },
    mTLS: {
      enabled: true,
      certificateBoundAccessTokens: true,
      selfSignedTlsClientAuth: true,
      tlsClientAuth: true,
      getCertificate(ctx) {
        const cert = ctx.req.socket.getPeerCertificate(true);
         console.log('Received client certificate:', JSON.stringify(cert, null, 2));
        if (!cert || Object.keys(cert).length === 0) {
          console.error('No client certificate provided or certificate is empty.');
          return null;
        }
        return cert;
      },
      certificateAuthorized(ctx) {
        const cert = ctx.req.socket.getPeerCertificate(true);
        console.log('Client certificate authorized:', cert);
        return cert && Object.keys(cert).length > 0;
      },
      certificateSubjectMatches(ctx, property, expected) {
        const cert = ctx.req.socket.getPeerCertificate(true);
        console.log(`Matching certificate ${property}:`, cert.subject[property]);
        console.log('Expected value:', expected);
        return cert.subject[property] === expected;
      },
    },
    fapi: {
      enabled: true,
      profile: '2.0',
    },
    jwtResponseModes: { enabled: true },
    revocation: { enabled: true },
  },
  clientAuthMethods: [
    'private_key_jwt',
  ],
  scopes: ['api:access'],
  clientDefaults: {
    grant_types: ['client_credentials'], 
    id_token_signed_response_alg: 'RS256',
    token_endpoint_auth_method: 'private_key_jwt',
  },
  issueRefreshToken: async (ctx, client, code) => {
    return client.grantTypes.includes('refresh_token');
  },
  ttl: {
    AccessToken: 60 * 60, // 1 hour
    RefreshToken: 14 * 24 * 60 * 60, // 14 days
    ClientCredentials: 10 * 60, // 10 minutes
  },
  rotateRefreshToken: true,
  allowOmittingSingleRegisteredRedirectUri: true,
  extraClientMetadata: {
    properties: ['software_statement', 'software_id', 'software_version'],
  },
};

export default oidcConfiguration;