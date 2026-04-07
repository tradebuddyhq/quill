package stdlib

// SecureStorageRuntime provides encrypted browser-side storage.
const SecureStorageRuntime = `
// Quill Secure Storage Runtime (browser)
// Uses SubtleCrypto for AES-GCM encryption on top of localStorage

const SecureStorage = {
  _key: null,

  async init(password) {
    const enc = new TextEncoder();
    const keyMaterial = await crypto.subtle.importKey(
      "raw", enc.encode(password), "PBKDF2", false, ["deriveKey"]
    );
    this._key = await crypto.subtle.deriveKey(
      { name: "PBKDF2", salt: enc.encode("quill-secure"), iterations: 100000, hash: "SHA-256" },
      keyMaterial, { name: "AES-GCM", length: 256 }, false, ["encrypt", "decrypt"]
    );
  },

  async set(name, value) {
    if (!this._key) throw new Error("SecureStorage not initialized. Call SecureStorage.init(password) first.");
    const enc = new TextEncoder();
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const encrypted = await crypto.subtle.encrypt(
      { name: "AES-GCM", iv }, this._key, enc.encode(JSON.stringify(value))
    );
    const data = { iv: Array.from(iv), data: Array.from(new Uint8Array(encrypted)) };
    localStorage.setItem("__qs_" + name, JSON.stringify(data));
  },

  async get(name) {
    if (!this._key) throw new Error("SecureStorage not initialized. Call SecureStorage.init(password) first.");
    const raw = localStorage.getItem("__qs_" + name);
    if (!raw) return null;
    const { iv, data } = JSON.parse(raw);
    const decrypted = await crypto.subtle.decrypt(
      { name: "AES-GCM", iv: new Uint8Array(iv) }, this._key, new Uint8Array(data)
    );
    return JSON.parse(new TextDecoder().decode(decrypted));
  },

  remove(name) {
    localStorage.removeItem("__qs_" + name);
  },

  clear() {
    const keys = Object.keys(localStorage).filter(k => k.startsWith("__qs_"));
    keys.forEach(k => localStorage.removeItem(k));
  }
};
`

// GetSecureStorageRuntime returns the secure storage runtime string.
func GetSecureStorageRuntime() string {
	return SecureStorageRuntime
}
