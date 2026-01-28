/**
 * Mock HuggingFace API server for testing
 * Serves responses from fixtures/moderation_responses.json
 */

const http = require('http');
const fs = require('fs');
const path = require('path');

const PORT = process.env.PORT || 8090;
const FIXTURES_PATH = process.env.FIXTURES_PATH || path.join(__dirname, '../fixtures/moderation_responses.json');

// Load fixture responses
let responses = {};
try {
  const data = fs.readFileSync(FIXTURES_PATH, 'utf-8');
  responses = JSON.parse(data);
  console.log(`Loaded ${Object.keys(responses).length} fixture responses`);
} catch (err) {
  console.error('Failed to load fixtures:', err);
  process.exit(1);
}

const server = http.createServer((req, res) => {
  // Handle CORS
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');

  if (req.method === 'OPTIONS') {
    res.writeHead(200);
    res.end();
    return;
  }

  // Only handle POST requests to model endpoint
  if (req.method !== 'POST') {
    res.writeHead(404, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ error: 'Not Found' }));
    return;
  }

  let body = '';
  req.on('data', chunk => {
    body += chunk.toString();
  });

  req.on('end', () => {
    try {
      const requestData = JSON.parse(body);
      const inputText = requestData.inputs || '';

      console.log(`Received request with input: "${inputText.substring(0, 50)}..."`);

      // Find matching response
      let matchedResponse = null;
      for (const [key, responseData] of Object.entries(responses)) {
        if (responseData.input === 'any text' || inputText.includes(responseData.input)) {
          matchedResponse = responseData;
          console.log(`Matched fixture: ${key}`);
          break;
        }
      }

      // Default to safe_text if no match
      if (!matchedResponse) {
        matchedResponse = responses.safe_text;
        console.log('No match found, using safe_text default');
      }

      // Handle error responses
      if (matchedResponse.error) {
        const status = matchedResponse.status || 500;
        res.writeHead(status, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ error: matchedResponse.error }));
        return;
      }

      // Return successful response
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify(matchedResponse.response));

    } catch (err) {
      console.error('Error processing request:', err);
      res.writeHead(400, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: 'Invalid request body' }));
    }
  });
});

server.listen(PORT, () => {
  console.log(`Mock HuggingFace API server running on port ${PORT}`);
  console.log(`Fixtures loaded from: ${FIXTURES_PATH}`);
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('SIGTERM received, closing server...');
  server.close(() => {
    console.log('Server closed');
    process.exit(0);
  });
});
