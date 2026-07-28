const assert = require('assert');
const crypto = require('crypto');
const http = require('http');
const { sendWebhook } = require('./index');

async function testWebhookHMAC() {
  const secret = 'zippyra-lambda-webhook-secret-32bytes';
  let receivedPayload = null;
  let receivedSignature = null;

  // Start mock catalog-service webhook HTTP server
  const server = http.createServer((req, res) => {
    let body = '';
    req.on('data', (chunk) => { body += chunk; });
    req.on('end', () => {
      receivedPayload = JSON.parse(body);
      receivedSignature = req.headers['x-signature'];

      const expectedSig = crypto
        .createHmac('sha256', secret)
        .update(body)
        .digest('hex');

      if (receivedSignature === expectedSig) {
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ status: 'success' }));
      } else {
        res.writeHead(401, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ error: 'invalid signature' }));
      }
    });
  });

  await new Promise((resolve) => server.listen(8083, '127.0.0.1', resolve));

  try {
    process.env.CATALOG_SERVICE_URL = 'http://127.0.0.1:8083';
    process.env.LAMBDA_WEBHOOK_SHARED_SECRET = secret;

    await sendWebhook({
      s3_raw_key: 'raw/prod_123.jpg',
      thumbnail_url: 'https://cdn.zippyra.com/thumbnails/prod_123.webp',
      full_url: 'https://cdn.zippyra.com/full/prod_123.webp',
      status: 'PROCESSED',
    });

    assert.strictEqual(receivedPayload.s3_raw_key, 'raw/prod_123.jpg');
    assert.strictEqual(receivedPayload.status, 'PROCESSED');
    assert.strictEqual(receivedPayload.thumbnail_url, 'https://cdn.zippyra.com/thumbnails/prod_123.webp');
    assert(receivedSignature, 'Signature header should be present');
    console.log('PASS: Lambda HMAC webhook unit test passed successfully.');
  } finally {
    server.close();
  }
}

testWebhookHMAC().catch((err) => {
  console.error('FAIL: Image processor unit test error:', err);
  process.exit(1);
});
