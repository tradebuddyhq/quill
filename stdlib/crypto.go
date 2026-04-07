package stdlib

// CryptoRuntime provides built-in cryptography functions.
const CryptoRuntime = `
// Quill Crypto Runtime
const __crypto = require("crypto");

function hash(data, algorithm) {
  const algo = algorithm || "sha256";
  return __crypto.createHash(algo).update(String(data)).digest("hex");
}

function hmac(data, key, algorithm) {
  const algo = algorithm || "sha256";
  return __crypto.createHmac(algo, key).update(String(data)).digest("hex");
}

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

function generateKeys() {
  const { publicKey, privateKey } = __crypto.generateKeyPairSync("rsa", {
    modulusLength: 2048,
    publicKeyEncoding: { type: "spki", format: "pem" },
    privateKeyEncoding: { type: "pkcs8", format: "pem" }
  });
  return { publicKey, privateKey };
}

function randomBytes(size) {
  return __crypto.randomBytes(size || 32).toString("hex");
}

function uuid() {
  return __crypto.randomUUID();
}
`

// GetCryptoRuntime returns the crypto runtime string.
func GetCryptoRuntime() string {
	return CryptoRuntime
}
