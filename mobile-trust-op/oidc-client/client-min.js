const https = require('https');
const fs = require('fs');

const clientCert = fs.readFileSync('../certificates/client.crt');
const clientKey = fs.readFileSync('../certificates/client.key');
const caCert = fs.readFileSync('../certificates/ca.crt');

const options = {
  hostname: 'localhost',
  port: 3000,
  path: '/token',
  method: 'GET',
  key: clientKey,
  cert: clientCert,
  ca: caCert,
  rejectUnauthorized: true,
};

const req = https.request(options, (res) => {
  console.log('statusCode:', res.statusCode);
  res.on('data', (d) => {
    process.stdout.write(d);
  });
});

req.on('error', (e) => {
  console.error('Error:', e);
});

req.end();