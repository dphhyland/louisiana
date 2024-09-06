import jwt from 'jsonwebtoken'; // Import JWT library
import fs from 'fs'; // Import file system module
import https from 'https'; // Import HTTPS module
import path from 'path'; // Import path module
import dotenv from 'dotenv'; // Import dotenv for environment variable support
import { fileURLToPath } from 'url'; // Import utilities to manage file paths

// Handle __dirname and __filename for .mjs files
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Load environment variables from .env file
dotenv.config();

// Load private key for signing JWT
const privateKey = fs.readFileSync(path.join(__dirname, '../certificates/client.key'));

// Function to create a JWT for client authentication
function createClientJWT(clientId, issuer, audience) {
  const now = Math.floor(Date.now() / 1000);
  const payload = {
    iss: clientId,
    sub: clientId,
    aud: audience, // The token endpoint URL
    jti: Math.random().toString(36).substring(2), // Unique identifier for the JWT
    iat: now,
    exp: now + 60, // JWT expiration (e.g., 60 seconds from now)
  };

  // Sign the JWT using the private key
  const token = jwt.sign(payload, privateKey, { algorithm: 'RS256', keyid: '6b1b52d1-98b4-46be-8999-6b6be144d21f' });

  return token;
}

// Create JWT for authentication
const clientId = 'localhost2';
const issuer = 'https://localhost';
const audience = 'https://localhost:3000/token'; // Token endpoint
const clientJWT = createClientJWT(clientId, issuer, audience);

// Log the JWT
console.log('Client JWT:', clientJWT);

// HTTPS agent for making requests
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // Set to true in production
});

// Data for token request
const data = `grant_type=client_credentials&client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer&client_assertion=${encodeURIComponent(clientJWT)}`;

// Options for HTTPS request
const options = {
  hostname: 'localhost',
  port: 3000,
  path: '/token',
  method: 'POST',
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded',
    'Content-Length': Buffer.byteLength(data), // Ensure correct content length
  },
  agent: httpsAgent,
};

// Send the HTTPS request
const req = https.request(options, (res) => {
  let responseData = '';
  res.on('data', (chunk) => {
    responseData += chunk;
  });
  res.on('end', () => {
    console.log('Response:', responseData); // Output the server response
  });
});

req.on('error', (e) => {
  console.error('Request error:', e); // Log any request errors
});

// Write data and end the request
req.write(data);
req.end();