import fs from 'fs';
import path from 'path';
import { fileURLToPath } from "url";
import { importPKCS8, exportJWK } from 'jose';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Load the private key PEM file
const privateKeyPEM = fs.readFileSync(path.join(__dirname, '../certificates/client.key'), 'utf8');

// Convert the PEM-formatted private key into a JWK
async function generateJwks() {
  try {
    // Import the PEM private key and convert to JWK using the RS256 algorithm
    const privateKey = await importPKCS8(privateKeyPEM, 'RS256');

    // Export the JWK (JSON Web Key)
    const privateKeyJWK = await exportJWK(privateKey);

    // Create a JWKS object containing the private key
    const jwks = {
      keys: [privateKeyJWK], // Add multiple keys if needed
    };

    console.log('Generated JWKS:', JSON.stringify(jwks, null, 2));

    // Optionally, write the JWKS to a file for future use
    fs.writeFileSync(path.join(__dirname, '../certificates/client.jwks.json'), JSON.stringify(jwks, null, 2));

    return jwks;
  } catch (error) {
    console.error('Error generating JWKS:', error);
  }
}

// Call the function to generate JWKS
generateJwks();