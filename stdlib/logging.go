package stdlib

// GetLoggingRuntime returns the JavaScript runtime for the Quill Log library.
// It provides structured logging with levels, JSON output, colored console output,
// child loggers, and request logging middleware.
func GetLoggingRuntime() string {
	return LoggingRuntime
}

const LoggingRuntime = `
// Quill Logging Runtime

(function(global) {
  "use strict";

  var LEVELS = { debug: 0, info: 1, warn: 2, error: 3, fatal: 4 };

  var LEVEL_NAMES = ["DEBUG", "INFO", "WARN", "ERROR", "FATAL"];

  var COLORS = {
    debug: "\x1b[36m",  // cyan
    info:  "\x1b[32m",  // green
    warn:  "\x1b[33m",  // yellow
    error: "\x1b[31m",  // red
    fatal: "\x1b[35m",  // magenta
    reset: "\x1b[0m",
    dim:   "\x1b[2m",
    bold:  "\x1b[1m"
  };

  function _timestamp() {
    return new Date().toISOString();
  }

  function _getCallerInfo() {
    try {
      var err = new Error();
      var stack = err.stack.split("\n");
      // Walk up the stack past internal logger frames
      for (var i = 3; i < stack.length; i++) {
        var line = stack[i];
        if (line && line.indexOf("Logger.") === -1 && line.indexOf("_log") === -1) {
          var match = line.match(/(?:at\s+)?(?:.*?\s+\()?([^()]+):(\d+):\d+\)?/);
          if (match) {
            return { file: match[1].trim(), line: parseInt(match[2], 10) };
          }
        }
      }
    } catch (e) {}
    return { file: "unknown", line: 0 };
  }

  // --- Logger class ---

  function Logger(options) {
    options = options || {};
    this.name = options.name || "app";
    this.level = options.level || "debug";
    this.format = options.format || "pretty"; // "pretty" or "json"
    this.context = options.context || {};
    this.timestamps = options.timestamps !== false;
    this.colorize = options.colorize !== false;
    this.source = options.source !== false;
  }

  Logger.prototype._shouldLog = function(level) {
    return LEVELS[level] >= LEVELS[this.level];
  };

  Logger.prototype._log = function(level, msg, data) {
    if (!this._shouldLog(level)) return;

    var entry = {
      timestamp: this.timestamps ? _timestamp() : undefined,
      level: level.toUpperCase(),
      logger: this.name,
      message: msg
    };

    if (this.source) {
      var caller = _getCallerInfo();
      entry.source = caller.file + ":" + caller.line;
    }

    // Merge context
    for (var k in this.context) {
      entry[k] = this.context[k];
    }

    // Merge additional data
    if (data && typeof data === "object") {
      for (var k in data) {
        entry[k] = data[k];
      }
    }

    if (this.format === "json") {
      this._outputJSON(entry, level);
    } else {
      this._outputPretty(entry, level);
    }

    if (level === "fatal") {
      if (typeof process !== "undefined" && process.exit) {
        process.exit(1);
      }
    }
  };

  Logger.prototype._outputJSON = function(entry, level) {
    var output = JSON.stringify(entry);
    if (level === "error" || level === "fatal") {
      console.error(output);
    } else if (level === "warn") {
      console.warn(output);
    } else {
      console.log(output);
    }
  };

  Logger.prototype._outputPretty = function(entry, level) {
    var parts = [];

    if (this.colorize) {
      var color = COLORS[level] || COLORS.reset;
      if (entry.timestamp) {
        parts.push(COLORS.dim + entry.timestamp + COLORS.reset);
      }
      parts.push(color + COLORS.bold + entry.level.padEnd(5) + COLORS.reset);
      parts.push(COLORS.dim + "[" + entry.logger + "]" + COLORS.reset);
      parts.push(entry.message);
      if (entry.source) {
        parts.push(COLORS.dim + "(" + entry.source + ")" + COLORS.reset);
      }
    } else {
      if (entry.timestamp) {
        parts.push(entry.timestamp);
      }
      parts.push(entry.level.padEnd(5));
      parts.push("[" + entry.logger + "]");
      parts.push(entry.message);
      if (entry.source) {
        parts.push("(" + entry.source + ")");
      }
    }

    var output = parts.join(" ");

    // Append data fields (excluding standard fields)
    var extras = {};
    var hasExtras = false;
    var standard = { timestamp: 1, level: 1, logger: 1, message: 1, source: 1 };
    for (var k in entry) {
      if (!standard[k]) {
        extras[k] = entry[k];
        hasExtras = true;
      }
    }
    if (hasExtras) {
      output += " " + JSON.stringify(extras);
    }

    if (level === "error" || level === "fatal") {
      console.error(output);
    } else if (level === "warn") {
      console.warn(output);
    } else {
      console.log(output);
    }
  };

  Logger.prototype.debug = function(msg, data) { this._log("debug", msg, data); };
  Logger.prototype.info  = function(msg, data) { this._log("info",  msg, data); };
  Logger.prototype.warn  = function(msg, data) { this._log("warn",  msg, data); };
  Logger.prototype.error = function(msg, data) { this._log("error", msg, data); };
  Logger.prototype.fatal = function(msg, data) { this._log("fatal", msg, data); };

  Logger.prototype.child = function(extraContext) {
    var childCtx = {};
    for (var k in this.context) childCtx[k] = this.context[k];
    for (var k in extraContext) childCtx[k] = extraContext[k];
    return new Logger({
      name: this.name,
      level: this.level,
      format: this.format,
      context: childCtx,
      timestamps: this.timestamps,
      colorize: this.colorize,
      source: this.source
    });
  };

  // --- Log API ---

  var Log = {};

  Log.create = function(options) {
    return new Logger(options);
  };

  Log.middleware = function(options) {
    options = options || {};
    var logger = options.logger || new Logger({ name: "http", format: options.format || "pretty" });
    return function(req, res, next) {
      var start = Date.now();
      var method = req.method;
      var path = req.url || req.path;

      // Hook into response finish
      var origEnd = res.end;
      res.end = function() {
        var duration = Date.now() - start;
        var status = res.statusCode;
        var level = status >= 500 ? "error" : (status >= 400 ? "warn" : "info");
        logger._log(level, method + " " + path, {
          status: status,
          duration: duration + "ms",
          method: method,
          path: path
        });
        origEnd.apply(res, arguments);
      };

      if (typeof next === "function") next();
    };
  };

  global.Log = Log;
  global.Logger = Logger;

})(typeof window !== "undefined" ? window : (typeof global !== "undefined" ? global : this));
`
