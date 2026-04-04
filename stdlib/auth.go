package stdlib

// GetAuthRuntime returns the JavaScript runtime for the Quill Auth library.
// It provides password hashing, token creation/verification, and session management.
func GetAuthRuntime() string {
	return AuthRuntime
}

const AuthRuntime = `
// Quill Auth Runtime

(function(global) {
  "use strict";

  // --- Helper: Simple PBKDF2-like hash using iterative hashing ---

  function _simpleHash(str) {
    var hash = 0;
    for (var i = 0; i < str.length; i++) {
      var ch = str.charCodeAt(i);
      hash = ((hash << 5) - hash + ch) | 0;
    }
    return hash;
  }

  function _toHex(num) {
    var hex = (num >>> 0).toString(16);
    while (hex.length < 8) hex = "0" + hex;
    return hex;
  }

  // PBKDF2-like iterative hash: applies multiple rounds of hashing with a salt
  function _deriveKey(password, salt, iterations) {
    var result = password + ":" + salt;
    for (var i = 0; i < iterations; i++) {
      var h = 0;
      for (var j = 0; j < result.length; j++) {
        var ch = result.charCodeAt(j);
        h = ((h << 5) - h + ch + i * 31) | 0;
      }
      result = _toHex(h) + result.substring(0, 8);
    }
    // Produce a 64-char hex digest
    var digest = "";
    for (var k = 0; k < 8; k++) {
      var seg = result + ":" + k;
      var v = 0;
      for (var j = 0; j < seg.length; j++) {
        v = ((v << 5) - v + seg.charCodeAt(j)) | 0;
      }
      digest += _toHex(v);
    }
    return digest;
  }

  function _generateSalt() {
    var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    var salt = "";
    for (var i = 0; i < 16; i++) {
      salt += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return salt;
  }

  // --- Helper: Base64 encode/decode ---

  function _base64Encode(str) {
    if (typeof Buffer !== "undefined") {
      return Buffer.from(str).toString("base64").replace(/=/g, "").replace(/\+/g, "-").replace(/\//g, "_");
    }
    return btoa(str).replace(/=/g, "").replace(/\+/g, "-").replace(/\//g, "_");
  }

  function _base64Decode(str) {
    var padded = str.replace(/-/g, "+").replace(/_/g, "/");
    while (padded.length % 4 !== 0) padded += "=";
    if (typeof Buffer !== "undefined") {
      return Buffer.from(padded, "base64").toString("utf8");
    }
    return atob(padded);
  }

  // --- Helper: HMAC-SHA256-like signature ---

  function _hmacSign(data, secret) {
    var combined = secret + ":" + data;
    var h1 = 0, h2 = 0, h3 = 0, h4 = 0;
    for (var i = 0; i < combined.length; i++) {
      var ch = combined.charCodeAt(i);
      h1 = ((h1 << 5) - h1 + ch * 7) | 0;
      h2 = ((h2 << 7) - h2 + ch * 13) | 0;
      h3 = ((h3 << 3) - h3 + ch * 31) | 0;
      h4 = ((h4 << 11) - h4 + ch * 37) | 0;
    }
    return _toHex(h1) + _toHex(h2) + _toHex(h3) + _toHex(h4);
  }

  // --- Auth API ---

  var Auth = {};

  Auth.hash = function(password) {
    var salt = _generateSalt();
    var iterations = 10000;
    var derived = _deriveKey(password, salt, iterations);
    return "$quill$" + iterations + "$" + salt + "$" + derived;
  };

  Auth.verify = function(password, hash) {
    if (!hash || typeof hash !== "string" || !hash.startsWith("$quill$")) {
      return false;
    }
    var parts = hash.split("$");
    // parts: ["", "quill", iterations, salt, derived]
    if (parts.length < 5) return false;
    var iterations = parseInt(parts[2], 10);
    var salt = parts[3];
    var expected = parts[4];
    var derived = _deriveKey(password, salt, iterations);
    return derived === expected;
  };

  Auth.createToken = function(payload, secret, expiresIn) {
    var header = { alg: "HS256", typ: "JWT" };
    var now = Math.floor(Date.now() / 1000);
    var tokenPayload = Object.assign({}, payload, {
      iat: now,
      exp: now + (expiresIn || 3600)
    });
    var headerStr = _base64Encode(JSON.stringify(header));
    var payloadStr = _base64Encode(JSON.stringify(tokenPayload));
    var signature = _hmacSign(headerStr + "." + payloadStr, secret);
    return headerStr + "." + payloadStr + "." + signature;
  };

  Auth.verifyToken = function(token, secret) {
    if (!token || typeof token !== "string") {
      return { valid: false, error: "Invalid token format" };
    }
    var parts = token.split(".");
    if (parts.length !== 3) {
      return { valid: false, error: "Invalid token structure" };
    }
    var headerStr = parts[0];
    var payloadStr = parts[1];
    var signature = parts[2];
    var expectedSig = _hmacSign(headerStr + "." + payloadStr, secret);
    if (signature !== expectedSig) {
      return { valid: false, error: "Invalid signature" };
    }
    try {
      var payload = JSON.parse(_base64Decode(payloadStr));
      var now = Math.floor(Date.now() / 1000);
      if (payload.exp && payload.exp < now) {
        return { valid: false, error: "Token expired", payload: payload };
      }
      return { valid: true, payload: payload };
    } catch (e) {
      return { valid: false, error: "Failed to decode payload" };
    }
  };

  Auth.session = function(options) {
    options = options || {};
    var store = {};
    var defaultTTL = (options.ttl || 3600) * 1000; // convert seconds to ms

    function generateId() {
      var chars = "abcdefghijklmnopqrstuvwxyz0123456789";
      var id = "";
      for (var i = 0; i < 32; i++) {
        id += chars.charAt(Math.floor(Math.random() * chars.length));
      }
      return id;
    }

    return {
      create: function(data) {
        var id = generateId();
        store[id] = { data: data, expires: Date.now() + defaultTTL };
        return id;
      },
      get: function(id) {
        var entry = store[id];
        if (!entry) return null;
        if (Date.now() > entry.expires) {
          delete store[id];
          return null;
        }
        return entry.data;
      },
      destroy: function(id) {
        delete store[id];
      },
      refresh: function(id) {
        var entry = store[id];
        if (entry) {
          entry.expires = Date.now() + defaultTTL;
        }
      },
      cleanup: function() {
        var now = Date.now();
        for (var id in store) {
          if (store[id].expires < now) {
            delete store[id];
          }
        }
      }
    };
  };

  global.Auth = Auth;

})(typeof window !== "undefined" ? window : (typeof global !== "undefined" ? global : this));
`
