package stdlib

// CryptoRuntime provides built-in cryptography functions.
const CryptoRuntime = `
// Quill Crypto Runtime
const __crypto = require("crypto");

// --- Hashing ---

function hash(data, algorithm) {
  const algo = algorithm || "sha256";
  return __crypto.createHash(algo).update(String(data)).digest("hex");
}

function hmac(data, key, algorithm) {
  const algo = algorithm || "sha256";
  if (typeof key === "string") key = Buffer.from(key);
  return __crypto.createHmac(algo, key).update(typeof data === "string" ? data : Buffer.from(data)).digest("hex");
}

function hmacBuffer(data, key, algorithm) {
  const algo = algorithm || "sha256";
  if (typeof key === "string") key = Buffer.from(key);
  return __crypto.createHmac(algo, key).update(typeof data === "string" ? Buffer.from(data) : data).digest();
}

// --- HKDF (HMAC-based Key Derivation Function) ---

function hkdf(inputKey, salt, info, outputLength) {
  const algo = "sha256";
  const hashLen = 32;
  const ikm = typeof inputKey === "string" ? Buffer.from(inputKey, "hex") : Buffer.from(inputKey);
  const saltBuf = salt ? (typeof salt === "string" ? Buffer.from(salt, "hex") : Buffer.from(salt)) : Buffer.alloc(hashLen, 0);
  const infoBuf = info ? (typeof info === "string" ? Buffer.from(info) : Buffer.from(info)) : Buffer.alloc(0);
  const len = outputLength || 32;

  // Extract
  const prk = __crypto.createHmac(algo, saltBuf).update(ikm).digest();

  // Expand
  const n = Math.ceil(len / hashLen);
  const okm = [];
  let prev = Buffer.alloc(0);
  for (let i = 1; i <= n; i++) {
    prev = __crypto.createHmac(algo, prk).update(Buffer.concat([prev, infoBuf, Buffer.from([i])])).digest();
    okm.push(prev);
  }
  return Buffer.concat(okm).slice(0, len).toString("hex");
}

function hkdfBuffer(inputKey, salt, info, outputLength) {
  const hex = hkdf(inputKey, salt, info, outputLength);
  return Buffer.from(hex, "hex");
}

// --- Diffie-Hellman (X25519) ---

function generateX25519Keys() {
  const { publicKey, privateKey } = __crypto.generateKeyPairSync("x25519", {
    publicKeyEncoding: { type: "spki", format: "der" },
    privateKeyEncoding: { type: "pkcs8", format: "der" }
  });
  // Extract raw 32-byte keys from DER encoding
  return {
    publicKey: publicKey.slice(-32).toString("hex"),
    privateKey: privateKey.slice(-32).toString("hex")
  };
}

function diffieHellman(myPrivateKeyHex, theirPublicKeyHex) {
  const myPrivateRaw = Buffer.from(myPrivateKeyHex, "hex");
  const theirPublicRaw = Buffer.from(theirPublicKeyHex, "hex");

  // Wrap raw keys in DER format for Node.js crypto
  const pkcs8Prefix = Buffer.from("302e020100300506032b656e04220420", "hex");
  const spkiPrefix = Buffer.from("302a300506032b656e032100", "hex");

  const privateKeyObj = __crypto.createPrivateKey({
    key: Buffer.concat([pkcs8Prefix, myPrivateRaw]),
    format: "der",
    type: "pkcs8"
  });
  const publicKeyObj = __crypto.createPublicKey({
    key: Buffer.concat([spkiPrefix, theirPublicRaw]),
    format: "der",
    type: "spki"
  });

  return __crypto.diffieHellman({ privateKey: privateKeyObj, publicKey: publicKeyObj }).toString("hex");
}

// --- RSA Key Generation ---

function generateKeys() {
  const { publicKey, privateKey } = __crypto.generateKeyPairSync("rsa", {
    modulusLength: 2048,
    publicKeyEncoding: { type: "spki", format: "pem" },
    privateKeyEncoding: { type: "pkcs8", format: "pem" }
  });
  return { publicKey, privateKey };
}

// --- Argon2id / bcrypt Password Hashing ---

async function argon2(password, salt, options) {
  // Use scrypt as the memory-hard KDF (native to Node.js)
  // For true argon2id, users should npm install argon2
  const opts = options || {};
  const mem = opts.memory || 65536;
  const iter = opts.iterations || 3;
  const saltBuf = salt ? Buffer.from(salt) : __crypto.randomBytes(16);
  const cost = Math.max(16384, Math.floor(mem / 4));
  const keyLen = opts.keyLength || 32;
  return new Promise((resolve, reject) => {
    __crypto.scrypt(password, saltBuf, keyLen, { N: cost, r: 8, p: iter }, (err, derived) => {
      if (err) return reject(err);
      resolve(saltBuf.toString("hex") + ":" + derived.toString("hex"));
    });
  });
}

async function argon2Verify(hashed, password) {
  const [saltHex, hashHex] = hashed.split(":");
  const saltBuf = Buffer.from(saltHex, "hex");
  const keyLen = Buffer.from(hashHex, "hex").length;
  return new Promise((resolve, reject) => {
    __crypto.scrypt(password, saltBuf, keyLen, { N: 16384, r: 8, p: 3 }, (err, derived) => {
      if (err) return reject(err);
      resolve(constantTimeEqual(derived.toString("hex"), hashHex));
    });
  });
}

async function bcryptHash(password, rounds) {
  // Fallback: use scrypt with configurable cost
  const saltBuf = __crypto.randomBytes(16);
  const cost = Math.pow(2, rounds || 12);
  return new Promise((resolve, reject) => {
    __crypto.scrypt(password, saltBuf, 32, { N: cost, r: 8, p: 1 }, (err, derived) => {
      if (err) return reject(err);
      resolve(saltBuf.toString("hex") + ":" + derived.toString("hex"));
    });
  });
}

async function bcryptVerify(hashed, password) {
  return argon2Verify(hashed, password);
}

// --- Constant-Time Comparison ---

function constantTimeEqual(a, b) {
  const bufA = typeof a === "string" ? Buffer.from(a) : Buffer.from(a);
  const bufB = typeof b === "string" ? Buffer.from(b) : Buffer.from(b);
  if (bufA.length !== bufB.length) return false;
  return __crypto.timingSafeEqual(bufA, bufB);
}

// --- AES-256-GCM / AES-256-CBC ---

function aesEncrypt(plaintext, key, iv, mode) {
  const algo = (mode === "aes-256-cbc") ? "aes-256-cbc" : "aes-256-gcm";
  const keyBuf = typeof key === "string" ? Buffer.from(key, "hex") : Buffer.from(key);
  const ivBuf = iv ? (typeof iv === "string" ? Buffer.from(iv, "hex") : Buffer.from(iv)) : __crypto.randomBytes(algo === "aes-256-gcm" ? 12 : 16);
  const cipher = __crypto.createCipheriv(algo, keyBuf, ivBuf);
  let encrypted = cipher.update(typeof plaintext === "string" ? plaintext : Buffer.from(plaintext), null, "hex");
  encrypted += cipher.final("hex");
  if (algo === "aes-256-gcm") {
    const tag = cipher.getAuthTag().toString("hex");
    return { ciphertext: encrypted, iv: ivBuf.toString("hex"), tag: tag };
  }
  return { ciphertext: encrypted, iv: ivBuf.toString("hex") };
}

function aesDecrypt(ciphertext, key, iv, tagOrMode, mode) {
  // aesDecrypt(ct, key, iv, tag, "aes-256-gcm") or aesDecrypt(ct, key, iv, "aes-256-cbc")
  let algo, tag;
  if (mode) {
    algo = mode;
    tag = tagOrMode;
  } else if (tagOrMode === "aes-256-cbc" || tagOrMode === "aes-256-gcm") {
    algo = tagOrMode;
    tag = null;
  } else {
    algo = "aes-256-gcm";
    tag = tagOrMode;
  }
  const keyBuf = typeof key === "string" ? Buffer.from(key, "hex") : Buffer.from(key);
  const ivBuf = typeof iv === "string" ? Buffer.from(iv, "hex") : Buffer.from(iv);
  const decipher = __crypto.createDecipheriv(algo, keyBuf, ivBuf);
  if (algo === "aes-256-gcm" && tag) {
    const tagBuf = typeof tag === "string" ? Buffer.from(tag, "hex") : Buffer.from(tag);
    decipher.setAuthTag(tagBuf);
  }
  let decrypted = decipher.update(ciphertext, "hex", "utf8");
  decrypted += decipher.final("utf8");
  return decrypted;
}

// --- Simple encrypt/decrypt (password-based) ---

function encrypt(text, password) {
  const key = __crypto.scryptSync(password, "quill-salt", 32);
  const iv = __crypto.randomBytes(16);
  const cipher = __crypto.createCipheriv("aes-256-gcm", key, iv);
  let encrypted = cipher.update(String(text), "utf8", "hex");
  encrypted += cipher.final("hex");
  const tag = cipher.getAuthTag().toString("hex");
  return iv.toString("hex") + ":" + tag + ":" + encrypted;
}

function decrypt(encrypted, password) {
  const [ivHex, tagHex, data] = encrypted.split(":");
  const key = __crypto.scryptSync(password, "quill-salt", 32);
  const iv = Buffer.from(ivHex, "hex");
  const tag = Buffer.from(tagHex, "hex");
  const decipher = __crypto.createDecipheriv("aes-256-gcm", key, iv);
  decipher.setAuthTag(tag);
  let decrypted = decipher.update(data, "hex", "utf8");
  decrypted += decipher.final("utf8");
  return decrypted;
}

// --- Random ---

function randomBytes(size) {
  return __crypto.randomBytes(size || 32).toString("hex");
}

function randomBytesBuffer(size) {
  return __crypto.randomBytes(size || 32);
}

function secureRandomInt(min, max) {
  const range = max - min;
  const bytesNeeded = Math.ceil(Math.log2(range) / 8) || 1;
  let randomValue;
  do {
    randomValue = parseInt(__crypto.randomBytes(bytesNeeded).toString("hex"), 16);
  } while (randomValue >= Math.floor(256 ** bytesNeeded / range) * range);
  return min + (randomValue % range);
}

function uuid() {
  return __crypto.randomUUID();
}

// --- Secure Erase ---

function secureErase(buffer) {
  if (Buffer.isBuffer(buffer)) {
    __crypto.randomFillSync(buffer);
    buffer.fill(0);
    return true;
  }
  if (typeof buffer === "string") {
    // Can't truly erase strings in JS (immutable), but return empty
    return "";
  }
  if (ArrayBuffer.isView(buffer)) {
    __crypto.randomFillSync(buffer);
    new Uint8Array(buffer.buffer).fill(0);
    return true;
  }
  return false;
}
`

// GetCryptoRuntime returns the crypto runtime string.
func GetCryptoRuntime() string {
	return CryptoRuntime
}
