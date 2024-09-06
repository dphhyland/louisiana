import { Provider } from 'oidc-provider';
import https from 'https';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import dotenv from 'dotenv';
import oidcConfiguration from './oidcConfig.mjs'; // Import configuration

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Load environment variables
dotenv.config();

// Load SSL certificates
const serverOptions = {
  key: fs.readFileSync(path.join(__dirname, '../certificates/server.key')),
  cert: fs.readFileSync(path.join(__dirname, '../certificates/server.crt')),
  ca: fs.readFileSync(path.join(__dirname, '../certificates/ca.crt')),
  rejectUnauthorized: false, // Don't reject unauthorized clients
};

const oidc = new Provider(`https://localhost:${process.env.PORT || 3000}`, oidcConfiguration);

// Middleware to log all requests
oidc.use(async (ctx, next) => {
  console.log(`Received ${ctx.method} request to ${ctx.path}`);
  console.log('Request headers:', ctx.headers);
  console.log('TLS authorized:', ctx.req.socket.authorized);
  console.log('TLS version:', ctx.req.socket.getProtocol());
  console.log('Cipher:', ctx.req.socket.getCipher());
  
  const cert = ctx.req.socket.getPeerCertificate(true);

  if (!cert || Object.keys(cert).length === 0) {
    console.error('No client certificate provided or certificate is empty.');
  } else {
    console.log('Client certificate received:');
    console.log('  Subject:', cert.subject);
    console.log('  Issuer:', cert.issuer);
    console.log('  Valid from:', cert.valid_from);
    console.log('  Valid to:', cert.valid_to);
  }

  await next();
});

// Middleware to log detailed response info
oidc.use(async (ctx, next) => {
  await next();
  console.log(`Response status: ${ctx.status}`);
  console.log('Response headers:', ctx.response.headers);
  console.log('Response body:', ctx.response.body);
});

const server = https.createServer(serverOptions, oidc.callback());

server.listen(process.env.PORT || 3000, () => {
  console.log(`oidc-provider listening on port ${process.env.PORT || 3000}`);
});

// Error handling
server.on('tlsClientError', (err, tlsSocket) => {
  console.error('TLS Client Error:', err);
  console.error('Error code:', err.code);
  console.error('Error message:', err.message);
});

server.on('clientError', (err, socket) => {
  console.error('Client Error:', err);
});

process.on('unhandledRejection', (reason, p) => {
  console.error('Unhandled Rejection at:', p, 'reason:', reason);
});