import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { importPKCS8, importSPKI, exportJWK } from 'jose';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Load the PEM files containing the keys
const privateKeyPEM = fs.readFileSync(path.join(__dirname, './certificates/server.key'), 'utf8');
const publicKeyPEM = fs.readFileSync(path.join(__dirname, './certificates/server_public_key.pem'), 'utf8');

async function convertPEMtoJWK() {
  try {
    // Import the PEM private key as a JWK object
    const privateKey = await importPKCS8(privateKeyPEM, 'RS256');
    const privateKeyJWK = await exportJWK(privateKey);

    // Output the private JWK
    console.log('Private JWK:', JSON.stringify(privateKeyJWK, null, 2));

    // Import the PEM public key as a JWK object
    const publicKey = await importSPKI(publicKeyPEM, 'RS256');
    const publicKeyJWK = await exportJWK(publicKey);

    // Output the public JWK
    console.log('Public JWK:', JSON.stringify(publicKeyJWK, null, 2));

  } catch (error) {
    console.error('Error converting PEM to JWK:', error);
  }
}

convertPEMtoJWK().catch((error) => console.error('Unhandled error:', error));