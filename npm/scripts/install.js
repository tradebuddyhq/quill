#!/usr/bin/env node
'use strict';

const https = require('https');
const fs = require('fs');
const path = require('path');

const REPO = 'tradebuddyhq/quill';
const BINARY_PATH = path.join(__dirname, '..', 'bin', 'quill-native');

function getPlatformKey() {
  const platform = process.platform;
  const arch = process.arch;

  const osMap = {
    darwin: 'darwin',
    linux: 'linux',
    win32: 'windows',
  };

  const archMap = {
    arm64: 'arm64',
    x64: 'amd64',
  };

  const os = osMap[platform];
  const cpu = archMap[arch];

  if (!os || !cpu) {
    return null;
  }

  const ext = platform === 'win32' ? '.exe' : '';
  return { name: `quill-${os}-${cpu}${ext}`, ext };
}

function httpsGet(url) {
  return new Promise((resolve, reject) => {
    https.get(url, { headers: { 'User-Agent': 'quill-npm-install' } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return httpsGet(res.headers.location).then(resolve, reject);
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`HTTP ${res.statusCode}`));
      }
      const chunks = [];
      res.on('data', (chunk) => chunks.push(chunk));
      res.on('end', () => resolve(Buffer.concat(chunks)));
      res.on('error', reject);
    }).on('error', reject);
  });
}

async function install() {
  const platformKey = getPlatformKey();
  if (!platformKey) {
    console.log('[quill] No prebuilt binary for this platform, using JS compiler.');
    return;
  }

  try {
    // Fetch latest release info
    const releaseData = await httpsGet(`https://api.github.com/repos/${REPO}/releases/latest`);
    const release = JSON.parse(releaseData.toString());

    // Find the matching asset
    const asset = release.assets.find((a) => a.name === platformKey.name);
    if (!asset) {
      console.log(`[quill] No binary found for ${platformKey.name}, using JS compiler.`);
      return;
    }

    // Download the binary
    console.log(`[quill] Downloading native binary (${platformKey.name})...`);
    const binary = await httpsGet(asset.browser_download_url);

    // Write to disk
    fs.writeFileSync(BINARY_PATH, binary);
    fs.chmodSync(BINARY_PATH, 0o755);

    console.log('[quill] Native binary installed successfully.');
  } catch (err) {
    console.log(`[quill] Could not download native binary: ${err.message}`);
    console.log('[quill] Falling back to JS compiler.');
  }
}

install();
