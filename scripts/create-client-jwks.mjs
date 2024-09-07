import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import crypto from 'crypto';
import { pem2jwk } from 'pem-jwk';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Load the client certificate
const certPem = fs.readFileSync(path.join(__dirname, '../certificates/client.crt'), 'utf8');

// Extract the public key from the certificate
const cert = new crypto.X509Certificate(certPem);
const publicKey = cert.publicKey;

// Convert the public key to JWK
const jwk = pem2jwk(publicKey.export({ type: 'spki', format: 'pem' }));

// Add additional properties
jwk.use = 'sig';
jwk.alg = 'RS256';
jwk.kid = crypto.createHash('sha256').update(certPem).digest('hex').slice(0, 16);

// Create the JWKS
const jwks = {
  keys: [jwk]
};

// Save the JWKS to a file
fs.writeFileSync(path.join(__dirname, '../certificates/client.jwks.json'), JSON.stringify(jwks, null, 2));

console.log('JWKS generated and saved.');
console.log(JSON.stringify(jwks, null, 2));