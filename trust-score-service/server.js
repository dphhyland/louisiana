const express = require('express');
const { exec } = require('child_process');
const bodyParser = require('body-parser');
const path = require('path');

const app = express();
app.use(bodyParser.json());

// Configuration
const config = {
  goProgramPath: process.env.GO_PROGRAM_PATH || './go-telephony-client',
  certFile: process.env.CERT_FILE || path.join(__dirname, 'certs', 'cert.crt'),
  keyFile: process.env.KEY_FILE || path.join(__dirname, 'certs', 'cert.key'),
  caFile: process.env.CA_FILE || path.join(__dirname, 'certs', 'ca.crt'),
  participantURL: process.env.PARTICIPANT_URL || 'https://data.sandbox.raidiam.io/participants',
  cacheFile: process.env.CACHE_FILE || path.join(__dirname, 'participants.json'),
  clientId: process.env.CLIENT_ID || 'https://rp.sandbox.raidiam.io/openid_relying_party/b683106b-126c-4577-9041-cb869de643a4'
};

function executeGoClient(inputJSON) {
    return new Promise((resolve, reject) => {
      const command = `${config.goProgramPath} -cert=${config.certFile} -key=${config.keyFile} -ca=${config.caFile} -participants=${config.participantURL} -cache=${config.cacheFile} -clientId=${config.clientId} '${inputJSON}'`;
      
      exec(command, (error, stdout, stderr) => {
        if (error) {
          console.error(`Error executing Go client: ${error.message}`);
          console.error(`Command: ${command}`);
          console.error(`Stdout: ${stdout}`);
          console.error(`Stderr: ${stderr}`);
          return reject(new Error('An error occurred while processing your request'));
        }
        if (stderr) {
          console.error(`Go client stderr: ${stderr}`);
        }
        
        // Extract JSON from stdout
        const jsonMatch = stdout.match(/\{[\s\S]*\}/);
        if (jsonMatch) {
          try {
            const result = JSON.parse(jsonMatch[0]);
            resolve(result);
          } catch (parseError) {
            console.error(`Error parsing Go client JSON output: ${parseError.message}`);
            console.error(`Matched JSON string: ${jsonMatch[0]}`);
            reject(new Error('Error processing the JSON response from the Go client'));
          }
        } else {
          console.error(`No valid JSON found in Go client output: ${stdout}`);
          reject(new Error('No valid JSON response from the Go client'));
        }
      });
    });
  }

  app.post('/check/telephony', async (req, res) => {
    try {
      const { mobile } = req.body;
      if (!mobile) {
        return res.status(400).json({ error: 'Mobile number is required' });
      }
  
      const inputJSON = JSON.stringify({ telephony: { mobile } });
      const result = await executeGoClient(inputJSON);
      res.json(result);
    } catch (error) {
      console.error(`Error in /check/telephony: ${error.message}`);
      if (error.message.includes('No valid JSON response')) {
        res.status(500).json({ 
          error: 'The Go client did not return a valid JSON response',
          details: error.message
        });
      } else if (error.message.includes('Error processing the JSON response')) {
        res.status(500).json({ 
          error: 'There was an error processing the JSON response from the Go client',
          details: error.message
        });
      } else {
        res.status(500).json({ 
          error: 'An unexpected error occurred while processing your request',
          details: error.message
        });
      }
    }
  });
  
  app.post('/check/bank', async (req, res) => {
    try {
      const { bsb, accountNumber, accountName } = req.body;
      if (!bsb || !accountNumber || !accountName) {
        return res.status(400).json({ error: 'BSB, account number, and account name are required' });
      }
  
      const inputJSON = JSON.stringify({ bank: { bsb, accountNumber, accountName } });
      const result = await executeGoClient(inputJSON);
      res.json(result);
    } catch (error) {
      console.error(`Error in /check/bank: ${error.message}`);
      res.status(500).json({ 
        error: 'An error occurred while processing your request',
        details: error.message
      });
    }
  });

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Express server is running on http://localhost:${PORT}`);
  console.log('Configuration:', config);
});