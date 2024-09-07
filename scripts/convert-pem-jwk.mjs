import fs from 'fs';
import { pem2jwk } from 'pem-jwk';

// Read the public key from PEM file
const pem = fs.readFileSync('./certificates/public_key.pem', 'utf-8');

// Convert PEM to JWK
const jwk = pem2jwk(pem);

// Output the JWK
console.log(JSON.stringify(jwk, null, 2));