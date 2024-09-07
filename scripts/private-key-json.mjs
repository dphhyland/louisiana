import { readFileSync, writeFileSync } from 'fs';
import { exportJWK, importPKCS8 } from 'jose';

const main = async () => {
  // Read the PEM-encoded private key
  const privateKeyPem = readFileSync('../certificates/client.key', 'utf8');

  // Import the PEM key as an RSA key in PKCS#8 format
  const privateKey = await importPKCS8(privateKeyPem, 'RS256');

  // Export the key to JWK format
  const jwk = await exportJWK(privateKey);

  // Save the JWK to a JSON file
  writeFileSync('../certificates/private-key.json', JSON.stringify(jwk, null, 2));
  console.log('Private key saved in JWK format as private-key.json');
};

main();