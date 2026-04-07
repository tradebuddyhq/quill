package stdlib

const WebRuntime = `
// Quill Web Server Runtime Helpers
// Provides convenience functions for building web servers with Express

const createServer = (options) => {
  const express = require("express");
  const app = express();
  // Parse JSON bodies by default
  app.use(express.json());
  // Parse URL-encoded bodies
  app.use(express.urlencoded({ extended: true }));
  if (options && options.static) {
    app.use(express.static(options.static));
  }
  return app;
};

const createSecureServer = (options) => {
  const express = require("express");
  const https = require("https");
  const fs = require("fs");
  const app = express();
  app.use(express.json());
  app.use(express.urlencoded({ extended: true }));
  if (options && options.static) {
    app.use(express.static(options.static));
  }
  const tlsOptions = {
    key: fs.readFileSync(options.key || "key.pem"),
    cert: fs.readFileSync(options.cert || "cert.pem")
  };
  if (options.ca) {
    tlsOptions.ca = fs.readFileSync(options.ca);
  }
  const server = https.createServer(tlsOptions, app);
  app.__httpsServer = server;
  const originalListen = app.listen.bind(app);
  app.listen = (port, cb) => {
    return server.listen(port, cb);
  };
  return app;
};
`
