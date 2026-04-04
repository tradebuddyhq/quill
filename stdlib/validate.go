package stdlib

// GetValidateRuntime returns the JavaScript runtime for the Quill Validate library.
// It provides schema-based validation, quick validators, and nested/array validation.
func GetValidateRuntime() string {
	return ValidateRuntime
}

const ValidateRuntime = `
// Quill Validate Runtime

(function(global) {
  "use strict";

  // --- Validation Rules ---

  function required(msg) {
    return { type: "required", message: msg || "This field is required" };
  }

  function minLength(n, msg) {
    return { type: "minLength", value: n, message: msg || ("Must be at least " + n + " characters") };
  }

  function maxLength(n, msg) {
    return { type: "maxLength", value: n, message: msg || ("Must be at most " + n + " characters") };
  }

  function min(n, msg) {
    return { type: "min", value: n, message: msg || ("Must be at least " + n) };
  }

  function max(n, msg) {
    return { type: "max", value: n, message: msg || ("Must be at most " + n) };
  }

  function email(msg) {
    return { type: "email", message: msg || "Must be a valid email address" };
  }

  function url(msg) {
    return { type: "url", message: msg || "Must be a valid URL" };
  }

  function pattern(regex, msg) {
    return { type: "pattern", value: regex, message: msg || "Does not match required pattern" };
  }

  function oneOf(values, msg) {
    return { type: "oneOf", value: values, message: msg || ("Must be one of: " + values.join(", ")) };
  }

  function custom(fn, msg) {
    return { type: "custom", value: fn, message: msg || "Custom validation failed" };
  }

  function arrayOf(schema, msg) {
    return { type: "arrayOf", value: schema, message: msg || "Invalid array element" };
  }

  // --- Email/URL regex patterns ---

  var EMAIL_RE = /^[a-zA-Z0-9.!#$%&'*+\/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;
  var URL_RE = /^https?:\/\/[^\s\/$.?#].[^\s]*$/i;

  // --- Validate a single value against a list of rules ---

  function validateField(value, rules, fieldPath) {
    var errors = [];
    for (var i = 0; i < rules.length; i++) {
      var rule = rules[i];
      switch (rule.type) {
        case "required":
          if (value === undefined || value === null || value === "") {
            errors.push({ field: fieldPath, message: rule.message });
          }
          break;
        case "minLength":
          if (typeof value === "string" && value.length < rule.value) {
            errors.push({ field: fieldPath, message: rule.message });
          }
          break;
        case "maxLength":
          if (typeof value === "string" && value.length > rule.value) {
            errors.push({ field: fieldPath, message: rule.message });
          }
          break;
        case "min":
          if (typeof value === "number" && value < rule.value) {
            errors.push({ field: fieldPath, message: rule.message });
          }
          break;
        case "max":
          if (typeof value === "number" && value > rule.value) {
            errors.push({ field: fieldPath, message: rule.message });
          }
          break;
        case "email":
          if (typeof value === "string" && !EMAIL_RE.test(value)) {
            errors.push({ field: fieldPath, message: rule.message });
          }
          break;
        case "url":
          if (typeof value === "string" && !URL_RE.test(value)) {
            errors.push({ field: fieldPath, message: rule.message });
          }
          break;
        case "pattern":
          var re = rule.value instanceof RegExp ? rule.value : new RegExp(rule.value);
          if (typeof value === "string" && !re.test(value)) {
            errors.push({ field: fieldPath, message: rule.message });
          }
          break;
        case "oneOf":
          if (rule.value.indexOf(value) === -1) {
            errors.push({ field: fieldPath, message: rule.message });
          }
          break;
        case "custom":
          if (!rule.value(value)) {
            errors.push({ field: fieldPath, message: rule.message });
          }
          break;
        case "arrayOf":
          if (Array.isArray(value)) {
            for (var j = 0; j < value.length; j++) {
              var subResult = validateData(value[j], rule.value, fieldPath + "[" + j + "]");
              for (var k = 0; k < subResult.length; k++) {
                errors.push(subResult[k]);
              }
            }
          } else if (value !== undefined && value !== null) {
            errors.push({ field: fieldPath, message: "Expected an array" });
          }
          break;
      }
    }
    return errors;
  }

  // --- Validate an entire data object against a schema ---

  function validateData(data, schemaDef, prefix) {
    var errors = [];
    prefix = prefix || "";
    for (var field in schemaDef) {
      var rules = schemaDef[field];
      var fieldPath = prefix ? prefix + "." + field : field;
      if (Array.isArray(rules)) {
        // Check if any rule is a nested schema (plain object without type property)
        var nestedSchema = null;
        var fieldRules = [];
        for (var i = 0; i < rules.length; i++) {
          if (rules[i] && typeof rules[i] === "object" && !rules[i].type) {
            nestedSchema = rules[i];
          } else {
            fieldRules.push(rules[i]);
          }
        }
        var value = data ? data[field] : undefined;
        var fieldErrors = validateField(value, fieldRules, fieldPath);
        for (var j = 0; j < fieldErrors.length; j++) {
          errors.push(fieldErrors[j]);
        }
        if (nestedSchema && value && typeof value === "object" && !Array.isArray(value)) {
          var nestedErrors = validateData(value, nestedSchema, fieldPath);
          for (var k = 0; k < nestedErrors.length; k++) {
            errors.push(nestedErrors[k]);
          }
        }
      } else if (typeof rules === "object" && !rules.type) {
        // Nested object schema
        var value = data ? data[field] : undefined;
        if (value && typeof value === "object") {
          var nestedErrors = validateData(value, rules, fieldPath);
          for (var k = 0; k < nestedErrors.length; k++) {
            errors.push(nestedErrors[k]);
          }
        }
      }
    }
    return errors;
  }

  // --- Schema wrapper ---

  function ValidationSchema(schemaDef) {
    this._schema = schemaDef;
  }

  ValidationSchema.prototype.check = function(data) {
    var errors = validateData(data, this._schema, "");
    return { valid: errors.length === 0, errors: errors };
  };

  // --- Validate API ---

  var Validate = {};

  Validate.schema = function(rules) {
    return new ValidationSchema(rules);
  };

  // Rule constructors
  Validate.required = required;
  Validate.minLength = minLength;
  Validate.maxLength = maxLength;
  Validate.min = min;
  Validate.max = max;
  Validate.email = email;
  Validate.url = url;
  Validate.pattern = pattern;
  Validate.oneOf = oneOf;
  Validate.custom = custom;
  Validate.arrayOf = arrayOf;

  // Quick validators
  Validate.isEmail = function(str) {
    return typeof str === "string" && EMAIL_RE.test(str);
  };

  Validate.isURL = function(str) {
    return typeof str === "string" && URL_RE.test(str);
  };

  Validate.isNumber = function(str) {
    if (typeof str === "number") return !isNaN(str);
    return typeof str === "string" && str.length > 0 && !isNaN(Number(str));
  };

  global.Validate = Validate;
  global.ValidationSchema = ValidationSchema;

})(typeof window !== "undefined" ? window : (typeof global !== "undefined" ? global : this));
`
