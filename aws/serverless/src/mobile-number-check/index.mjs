import express from 'express';
import jwt from 'jsonwebtoken';
import http from 'http';
import { calculateTrustScore } from './trustScoreService.mjs';
import { verifyToken } from './tokenUtils.mjs'; // Function to verify the JWT token
import dotenv from 'dotenv';
import fs from 'fs';
import path from 'path';

const __dirname = path.resolve();

// Load the private key from the file system
const privateKey = fs.readFileSync(path.resolve(__dirname, './certificates/server.key'), 'utf8');

// Load environment variables
dotenv.config();

const app = express();
app.use(express.json());

// Regular expression for E.164 phone number format
const e164Regex = /^\+?[1-9]\d{1,14}$/;

app.post('/trust-score', (req, res) => {

  const { mobile_number } = req.body;

  if (!mobile_number) {
    console.error('Mobile number is required.');  
    return res.status(400).json({ error: 'Mobile number is required.' });
  }

  // Validate phone number format against E.164 standard
  if (!e164Regex.test(mobile_number)) {
    console.error('Invalid phone number format. Must be in E.164 format.');
    return res.status(400).json({ error: 'Invalid phone number format. Must be in E.164 format.' });
  }

  // Calculate the trust score for the mobile number
  const trustScore = calculateTrustScore(mobile_number);

  // Generate the JWT payload for the API response
  const payload = {
    sub: mobile_number,
    score: trustScore,
    iss: process.env.ISSUER,
    iat: Math.floor(Date.now() / 1000), // Issued at time in seconds
  };

  // Sign the JWT with the server's private key using RS256 algorithm
  const signedJwt = jwt.sign(payload, privateKey, { algorithm: 'RS256' });

  // Respond with the signed JWT
  res.json({ jwt: signedJwt });
});

// Create an HTTPS server (no mTLS)
http.createServer( app).listen(process.env.PORT, () => {
  console.log(`Trust Score API listening on port ${process.env.PORT}`);

});