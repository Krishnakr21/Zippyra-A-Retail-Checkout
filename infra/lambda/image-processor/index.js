let s3Client;
try {
  const { S3Client } = require('@aws-sdk/client-s3');
  s3Client = new S3Client({ region: process.env.AWS_REGION || 'ap-south-1' });
} catch (e) {
  s3Client = null;
}

const sharp = (() => {
  try { return require('sharp'); } catch (e) { return null; }
})();
const crypto = require('crypto');
const http = require('http');
const https = require('https');
const { URL } = require('url');

function getCatalogServiceUrl() {
  return process.env.CATALOG_SERVICE_URL || 'http://catalog-service.internal:8083';
}
function getWebhookSecret() {
  return process.env.LAMBDA_WEBHOOK_SHARED_SECRET || 'zippyra-lambda-webhook-secret-32bytes';
}

async function handler(event) {
  for (const record of event.Records || []) {
    const bucket = record.s3.bucket.name;
    const rawKey = decodeURIComponent(record.s3.object.key.replace(/\+/g, ' '));

    // Guardrail against infinite loops: process only raw/ prefix
    if (!rawKey.startsWith('raw/')) {
      console.log(`Skipping non-raw object: ${rawKey}`);
      continue;
    }

    try {
      console.log(`Processing image for key: ${rawKey} in bucket: ${bucket}`);

      // 1. Download raw image from S3
      const getObj = await s3Client.send(new GetObjectCommand({ Bucket: bucket, Key: rawKey }));
      const inputBuffer = await streamToBuffer(getObj.Body);

      // 2. Process image with Sharp
      // Thumbnail: 400x400 WebP quality=80
      const thumbnailBuffer = await sharp(inputBuffer)
        .resize(400, 400, { fit: 'inside', withoutEnlargement: true })
        .webp({ quality: 80 })
        .toBuffer();

      // Full size: 800x800 WebP quality=80
      const fullBuffer = await sharp(inputBuffer)
        .resize(800, 800, { fit: 'inside', withoutEnlargement: true })
        .webp({ quality: 80 })
        .toBuffer();

      // 3. Define output keys
      const baseName = rawKey.replace(/^raw\//, '').replace(/\.[^.]+$/, '');
      const thumbnailKey = `thumbnails/${baseName}.webp`;
      const fullKey = `full/${baseName}.webp`;

      // 4. Upload processed images to S3
      await s3Client.send(new PutObjectCommand({
        Bucket: bucket,
        Key: thumbnailKey,
        Body: thumbnailBuffer,
        ContentType: 'image/webp',
      }));

      await s3Client.send(new PutObjectCommand({
        Bucket: bucket,
        Key: fullKey,
        Body: fullBuffer,
        ContentType: 'image/webp',
      }));

      const cloudfrontDomain = process.env.CLOUDFRONT_DOMAIN || `${bucket}.s3.amazonaws.com`;
      const thumbnailUrl = `https://${cloudfrontDomain}/${thumbnailKey}`;
      const fullUrl = `https://${cloudfrontDomain}/${fullKey}`;

      // 5. Notify catalog-service via HMAC-signed webhook
      await sendWebhook({
        s3_raw_key: rawKey,
        thumbnail_url: thumbnailUrl,
        full_url: fullUrl,
        status: 'PROCESSED',
      });

      console.log(`Successfully processed ${rawKey} -> ${thumbnailKey}, ${fullKey}`);
    } catch (err) {
      console.error(`Error processing image ${rawKey}:`, err);

      // Move corrupt/failed files to failed/ prefix
      try {
        const failedKey = `failed/${rawKey.replace(/^raw\//, '')}`;
        await s3Client.send(new PutObjectCommand({
          Bucket: bucket,
          Key: failedKey,
          Body: Buffer.from(`Processing failed: ${err.message}`),
          ContentType: 'text/plain',
        }));
      } catch (failedErr) {
        console.error(`Failed to write DLQ object for ${rawKey}:`, failedErr);
      }

      // Send failure notification webhook to catalog-service
      await sendWebhook({
        s3_raw_key: rawKey,
        status: 'FAILED',
        error_message: err.message || 'Image processing failed',
      });
    }
  }
}

function streamToBuffer(stream) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    stream.on('data', (chunk) => chunks.push(chunk));
    stream.on('error', reject);
    stream.on('end', () => resolve(Buffer.concat(chunks)));
  });
}

function sendWebhook(payload) {
  return new Promise((resolve, reject) => {
    const jsonBody = JSON.stringify(payload);
    const secret = getWebhookSecret();
    const catalogUrl = getCatalogServiceUrl();
    const signature = crypto
      .createHmac('sha256', secret)
      .update(jsonBody)
      .digest('hex');

    const targetUrl = new URL('/v1/catalog/internal/image-processed', catalogUrl);
    const transport = targetUrl.protocol === 'https:' ? https : http;

    const req = transport.request(targetUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Signature': signature,
        'Content-Length': Buffer.byteLength(jsonBody),
      },
    }, (res) => {
      let data = '';
      res.on('data', (chunk) => { data += chunk; });
      res.on('end', () => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(data);
        } else {
          reject(new Error(`Webhook failed with HTTP ${res.statusCode}: ${data}`));
        }
      });
    });

    req.on('error', reject);
    req.write(jsonBody);
    req.end();
  });
}

module.exports = {
  handler,
  sendWebhook,
  streamToBuffer,
};
