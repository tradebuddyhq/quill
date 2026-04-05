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
`
