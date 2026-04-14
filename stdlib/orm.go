package stdlib

// GetOrmRuntime returns the JavaScript runtime for the Quill ORM library.
// It provides database connectivity, model definition, migrations, and query building.
func GetOrmRuntime() string {
	return OrmRuntime
}

const OrmRuntime = `
// Quill ORM Runtime

(function(global) {
  "use strict";

  // --- Column name sanitization ---

  var __validColumn = /^[a-zA-Z_][a-zA-Z0-9_]*$/;
  function __safeCol(name) {
    if (!__validColumn.test(name)) {
      throw new Error("Invalid column name: " + name);
    }
    return name;
  }

  // --- Query Builder ---

  function QueryBuilder(model) {
    this._model = model;
    this._conditions = [];
    this._orderByField = null;
    this._orderDir = "ASC";
    this._limitVal = null;
    this._offsetVal = null;
  }

  QueryBuilder.prototype.where = function(field, op, value) {
    this._conditions.push({ field: field, op: op, value: value });
    return this;
  };

  QueryBuilder.prototype.orderBy = function(field, dir) {
    this._orderByField = field;
    this._orderDir = (dir && dir.toUpperCase() === "DESC") ? "DESC" : "ASC";
    return this;
  };

  QueryBuilder.prototype.limit = function(n) {
    this._limitVal = n;
    return this;
  };

  QueryBuilder.prototype.offset = function(n) {
    this._offsetVal = n;
    return this;
  };

  QueryBuilder.prototype._buildWhere = function() {
    if (this._conditions.length === 0) return { sql: "", params: [] };
    var clauses = [];
    var params = [];
    for (var i = 0; i < this._conditions.length; i++) {
      var c = this._conditions[i];
      clauses.push(__safeCol(c.field) + " " + c.op + " ?");
      params.push(c.value);
    }
    return { sql: " WHERE " + clauses.join(" AND "), params: params };
  };

  QueryBuilder.prototype._buildSuffix = function() {
    var sql = "";
    if (this._orderByField) {
      sql += " ORDER BY " + __safeCol(this._orderByField) + " " + this._orderDir;
    }
    if (this._limitVal !== null) {
      sql += " LIMIT " + this._limitVal;
    }
    if (this._offsetVal !== null) {
      sql += " OFFSET " + this._offsetVal;
    }
    return sql;
  };

  QueryBuilder.prototype.exec = function() {
    var w = this._buildWhere();
    var sql = "SELECT * FROM " + this._model._tableName + w.sql + this._buildSuffix();
    return this._model._db._execQuery(sql, w.params);
  };

  QueryBuilder.prototype.execOne = function() {
    this._limitVal = 1;
    var results = this.exec();
    return results.length > 0 ? results[0] : null;
  };

  // --- Type mapping ---

  var TYPE_MAP = {
    string: "TEXT",
    text: "TEXT",
    number: "REAL",
    integer: "INTEGER",
    int: "INTEGER",
    boolean: "INTEGER",
    bool: "INTEGER",
    date: "TEXT",
    datetime: "TEXT",
    json: "TEXT",
    blob: "BLOB"
  };

  function sqlType(typeName) {
    return TYPE_MAP[typeName.toLowerCase()] || "TEXT";
  }

  // --- Model ---

  function Model(db, name, schema) {
    this._db = db;
    this._name = name;
    this._tableName = __safeCol(name.toLowerCase() + "s");
    this._schema = schema;
    this._fields = Object.keys(schema);
  }

  Model.prototype.create = function(data) {
    var fields = [];
    var placeholders = [];
    var values = [];
    for (var i = 0; i < this._fields.length; i++) {
      var f = this._fields[i];
      if (data[f] !== undefined) {
        fields.push(__safeCol(f));
        placeholders.push("?");
        values.push(data[f]);
      }
    }
    var sql = "INSERT INTO " + this._tableName + " (" + fields.join(", ") + ") VALUES (" + placeholders.join(", ") + ")";
    return this._db._execRun(sql, values);
  };

  Model.prototype.find = function(query) {
    if (!query || Object.keys(query).length === 0) {
      return this._db._execQuery("SELECT * FROM " + this._tableName, []);
    }
    var clauses = [];
    var params = [];
    for (var key in query) {
      clauses.push(__safeCol(key) + " = ?");
      params.push(query[key]);
    }
    var sql = "SELECT * FROM " + this._tableName + " WHERE " + clauses.join(" AND ");
    return this._db._execQuery(sql, params);
  };

  Model.prototype.findOne = function(query) {
    var results = this.find(query);
    return results.length > 0 ? results[0] : null;
  };

  Model.prototype.update = function(query, data) {
    var setClauses = [];
    var params = [];
    for (var key in data) {
      setClauses.push(__safeCol(key) + " = ?");
      params.push(data[key]);
    }
    var whereClauses = [];
    for (var key in query) {
      whereClauses.push(__safeCol(key) + " = ?");
      params.push(query[key]);
    }
    var sql = "UPDATE " + this._tableName + " SET " + setClauses.join(", ");
    if (whereClauses.length > 0) {
      sql += " WHERE " + whereClauses.join(" AND ");
    }
    return this._db._execRun(sql, params);
  };

  Model.prototype.delete = function(query) {
    var clauses = [];
    var params = [];
    for (var key in query) {
      clauses.push(__safeCol(key) + " = ?");
      params.push(query[key]);
    }
    var sql = "DELETE FROM " + this._tableName;
    if (clauses.length > 0) {
      sql += " WHERE " + clauses.join(" AND ");
    }
    return this._db._execRun(sql, params);
  };

  Model.prototype.query = function() {
    return new QueryBuilder(this);
  };

  // --- Database ---

  var DB = {};
  DB._models = {};
  DB._connection = null;
  DB._driver = null;

  DB.connect = function(config) {
    config = config || {};
    var driver = config.driver || "sqlite";

    if (driver === "sqlite") {
      try {
        var Database = require("better-sqlite3");
        DB._connection = new Database(config.database || ":memory:");
        DB._driver = "sqlite";
        DB._execQuery = function(sql, params) {
          var stmt = DB._connection.prepare(sql);
          return stmt.all.apply(stmt, params || []);
        };
        DB._execRun = function(sql, params) {
          var stmt = DB._connection.prepare(sql);
          return stmt.run.apply(stmt, params || []);
        };
      } catch (e) {
        // Fallback: in-memory store if better-sqlite3 not available
        DB._driver = "memory";
        DB._memStore = {};
        DB._execQuery = function(sql, params) {
          // Basic in-memory query support
          var match = sql.match(/FROM\s+(\w+)/i);
          var table = match ? match[1] : null;
          if (!table || !DB._memStore[table]) return [];
          var rows = DB._memStore[table] || [];
          // Simple WHERE filtering
          var whereMatch = sql.match(/WHERE\s+(.+?)(?:\s+ORDER|\s+LIMIT|\s*$)/i);
          if (whereMatch && params && params.length > 0) {
            var idx = 0;
            rows = rows.filter(function(row) {
              var conditions = whereMatch[1].split(/\s+AND\s+/i);
              return conditions.every(function(cond) {
                var parts = cond.trim().split(/\s+/);
                var field = parts[0];
                var op = parts[1];
                var val = params[idx++];
                if (op === "=") return row[field] == val;
                if (op === ">") return row[field] > val;
                if (op === "<") return row[field] < val;
                if (op === ">=") return row[field] >= val;
                if (op === "<=") return row[field] <= val;
                if (op === "!=") return row[field] != val;
                return true;
              });
            });
          }
          return rows;
        };
        DB._execRun = function(sql, params) {
          var insertMatch = sql.match(/INSERT\s+INTO\s+(\w+)\s*\(([^)]+)\)\s*VALUES\s*\(([^)]+)\)/i);
          if (insertMatch) {
            var table = insertMatch[1];
            if (!DB._memStore[table]) DB._memStore[table] = [];
            var fields = insertMatch[2].split(",").map(function(f) { return f.trim(); });
            var row = { id: DB._memStore[table].length + 1 };
            for (var i = 0; i < fields.length; i++) {
              row[fields[i]] = params[i];
            }
            DB._memStore[table].push(row);
            return { changes: 1, lastInsertRowid: row.id };
          }
          return { changes: 0 };
        };
      }
    } else if (driver === "postgres" || driver === "pg") {
      var pg = require("pg");
      var pool = new pg.Pool(config);
      DB._connection = pool;
      DB._driver = "postgres";
      DB._execQuery = function(sql, params) {
        return pool.query(sql.replace(/\?/g, function() { return "$" + (++DB._pgIdx); }), params).then(function(res) { return res.rows; });
      };
      DB._execRun = function(sql, params) {
        DB._pgIdx = 0;
        return pool.query(sql.replace(/\?/g, function() { return "$" + (++DB._pgIdx); }), params).then(function(res) { return { changes: res.rowCount }; });
      };
      DB._pgIdx = 0;
    } else if (driver === "mysql") {
      var mysql = require("mysql2");
      var conn = mysql.createConnection(config);
      DB._connection = conn;
      DB._driver = "mysql";
      DB._execQuery = function(sql, params) {
        return new Promise(function(resolve, reject) {
          conn.query(sql, params, function(err, rows) {
            if (err) reject(err); else resolve(rows);
          });
        });
      };
      DB._execRun = function(sql, params) {
        return new Promise(function(resolve, reject) {
          conn.query(sql, params, function(err, result) {
            if (err) reject(err); else resolve({ changes: result.affectedRows, lastInsertRowid: result.insertId });
          });
        });
      };
    }

    return DB;
  };

  DB.model = function(name, schema) {
    var m = new Model(DB, name, schema);
    DB._models[name] = m;
    return m;
  };

  DB.migrate = function() {
    for (var name in DB._models) {
      var m = DB._models[name];
      var columns = ["id INTEGER PRIMARY KEY AUTOINCREMENT"];
      for (var field in m._schema) {
        var def = m._schema[field];
        var type = typeof def === "string" ? def : (def.type || "text");
        var col = __safeCol(field) + " " + sqlType(type);
        if (def.required) col += " NOT NULL";
        if (def.unique) col += " UNIQUE";
        if (def.default !== undefined) col += " DEFAULT " + JSON.stringify(def.default);
        columns.push(col);
      }
      var sql = "CREATE TABLE IF NOT EXISTS " + m._tableName + " (" + columns.join(", ") + ")";
      if (DB._driver === "memory") {
        if (!DB._memStore[m._tableName]) DB._memStore[m._tableName] = [];
      } else {
        DB._execRun(sql, []);
      }
    }
  };

  DB.raw = function(sql, params) {
    if (sql.trim().toUpperCase().startsWith("SELECT")) {
      return DB._execQuery(sql, params || []);
    }
    return DB._execRun(sql, params || []);
  };

  DB.transaction = function(callback) {
    if (DB._driver === "sqlite" && DB._connection) {
      var transaction = DB._connection.transaction(function() {
        callback(DB);
      });
      return transaction();
    } else if (DB._driver === "memory") {
      // Simple pass-through for in-memory
      return callback(DB);
    } else {
      // For async drivers, wrap in BEGIN/COMMIT
      return DB._execRun("BEGIN", []).then(function() {
        return callback(DB);
      }).then(function(result) {
        return DB._execRun("COMMIT", []).then(function() { return result; });
      }).catch(function(err) {
        return DB._execRun("ROLLBACK", []).then(function() { throw err; });
      });
    }
  };

  global.DB = DB;
  global.QueryBuilder = QueryBuilder;

})(typeof window !== "undefined" ? window : (typeof global !== "undefined" ? global : this));
`
