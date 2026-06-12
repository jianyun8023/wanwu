#!/usr/bin/env node
/*
 * 跨平台脚本：将 public_key.pem 转换为 public-key.js
 * 输出格式: window.__VEGA_PUBLIC_KEY__ = "<PEM 内容>";
 *
 * Forward-compatible: reads from data-connection/public_key.pem if the new
 * path doesn't have a non-empty file, and copies it over.
 *
 * 用法:
 *   node generate-public-key-js.js
 */
const fs = require('fs');
const path = require('path');

const baseDir = __dirname;
const pemPath = path.join(baseDir, 'public_key.pem');
const oldPemPath = path.join(baseDir, 'data-connection', 'public_key.pem');
const outPath = path.join(baseDir, 'public-key.js');

function isNonEmpty(p) {
  try { return fs.statSync(p).size > 0; } catch { return false; }
}

// Check output: refuse if non-empty
if (isNonEmpty(outPath)) {
  console.log(`Existing non-empty file detected, skipping generation: ${outPath}`);
  console.log('Delete or clear this file first if you want to regenerate.');
  process.exit(0);
}

// Resolve pem source: new path first, then old path for forward-compat
let sourcePem = null;
if (isNonEmpty(pemPath)) {
  sourcePem = pemPath;
} else if (isNonEmpty(oldPemPath)) {
  console.log(`Found public key in old path, copying: ${oldPemPath} -> ${pemPath}`);
  fs.copyFileSync(oldPemPath, pemPath);
  sourcePem = pemPath;
} else {
  console.error(`Public key not found or empty: ${pemPath}`);
  console.error('Run ./generate-keys.sh first.');
  process.exit(1);
}

const pem = fs.readFileSync(sourcePem, 'utf8').replace(/\r\n/g, '\n').trim();
const content = `window.__VEGA_PUBLIC_KEY__ = ${JSON.stringify(pem)};\n`;

fs.writeFileSync(outPath, content, { encoding: 'utf8' });

console.log(`Wrote ${outPath}`);
