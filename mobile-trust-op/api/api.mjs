import express from 'express';
import jwt from 'jsonwebtoken';
import https from 'https';
import fs from 'fs';
import path from 'path';
import { calculateTrustScore } from '../src/services/trustScoreService.mjs';
import { verifyToken } from '../src/utils/tokenUtils.mjs'; // Function to verify the JWT token

const __dirname = path.resolve();
const serverKey = fs.readFileSync(path.resolve(__dirname, './certificates/server.key'), 'utf8'); // Private key for the server
const serverCert = fs.readFileSync(path.resolve(__dirname, './certificates/server.crt')); // Public certificate for the server
const caCert = fs.readFileSync(path.resolve(__dirname, './certificates/ca.crt')); // CA certificate for client certificate verification

const app = express();
app.use(express.json());

const jwtIssuer = 'telcoexample.com'; // Issuer claim for the JWT

// Regular expression for E.164 phone number format
const e164Regex = /^\+?[1-9]\d{1,14}$/;

app.post('/api/trust-score', (req, res) => {
  // Extract client certificate
  const clientCert = req.connection.getPeerCertificate();
  
  if (!clientCert || Object.keys(clientCert).length === 0) {
    return res.status(401).json({ error: 'Unauthorized access - client certificate is required.' });
  }

  // Extract access token from the Authorization header
  const accessToken = req.headers.authorization?.split(' ')[1];
  if (!accessToken) {
    return res.status(401).json({ error: 'Unauthorized access - missing access token.' });
  }

  try {
    // Verify the access token using the Authorization Server's public key
    const decodedToken = verifyAccessToken(accessToken); // Make sure this function verifies the JWT signature and checks its validity

    // Check that the token was issued to the same client
    const tokenSubject = decodedToken.sub; // This should match a unique client identifier, such as the client certificate's CN or other field

    // For this example, let's assume the token subject matches the client certificate's subject CN
    if (clientCert.subject.CN !== tokenSubject) {
      return res.status(401).json({ error: 'Unauthorized access - token subject does not match client certificate.' });
    }

  } catch (err) {
    return res.status(401).json({ error: 'Unauthorized access - invalid token.' });
  }

  const { mobile_number } = req.body;
  if (!mobile_number) {
    return res.status(400).json({ error: 'Mobile number is required.' });
  }

  // Validate phone number format against E.164 standard
  if (!e164Regex.test(mobile_number)) {
    return res.status(400).json({ error: 'Invalid phone number format. Must be in E.164 format.' });
  }

  // Calculate the trust score for the mobile number
  const trustScore = calculateTrustScore(mobile_number);

  // Generate the JWT payload for the API response
  const payload = {
    sub: mobile_number,
    score: trustScore,
    iss: jwtIssuer,
    iat: Math.floor(Date.now() / 1000), // Issued at time in seconds
  };

  // Sign the JWT with the server's private key using RS256 algorithm
  const signedJwt = jwt.sign(payload, serverKey, { algorithm: 'RS256' });

  // Respond with the signed JWT
  res.json({ jwt: signedJwt });
});

https.createServer({
  key: serverKey,
  cert: serverCert,
  ca: caCert,
  requestCert: true,
  rejectUnauthorized: true,
}, app).listen(4000, () => {
  console.log('Trust Score API listening on port 4000 with mTLS enabled');
});