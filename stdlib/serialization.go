package stdlib

// SerializationRuntime provides binary encoding/decoding (lightweight protobuf-like).
const SerializationRuntime = `
// Quill Binary Serialization Runtime
// Compact binary encoding with schema support (no protobuf dependency)

const __BinaryTypes = {
  uint8: 1, uint16: 2, uint32: 3, int8: 4, int16: 5, int32: 6,
  float32: 7, float64: 8, string: 9, bytes: 10, bool: 11
};

function defineSchema(fields) {
  // fields: { name: { type: "string", tag: 1 }, age: { type: "uint8", tag: 2 } }
  return { fields, _isSchema: true };
}

function encode(schema, data) {
  const parts = [];
  for (const [name, spec] of Object.entries(schema.fields)) {
    const val = data[name];
    if (val === undefined || val === null) continue;
    const tag = spec.tag || 0;

    // Write tag + type byte
    parts.push(Buffer.from([tag]));
    parts.push(Buffer.from([__BinaryTypes[spec.type] || 0]));

    switch (spec.type) {
      case "uint8": parts.push(Buffer.from([val & 0xFF])); break;
      case "uint16": { const b = Buffer.alloc(2); b.writeUInt16BE(val); parts.push(b); break; }
      case "uint32": { const b = Buffer.alloc(4); b.writeUInt32BE(val); parts.push(b); break; }
      case "int8": { const b = Buffer.alloc(1); b.writeInt8(val); parts.push(b); break; }
      case "int16": { const b = Buffer.alloc(2); b.writeInt16BE(val); parts.push(b); break; }
      case "int32": { const b = Buffer.alloc(4); b.writeInt32BE(val); parts.push(b); break; }
      case "float32": { const b = Buffer.alloc(4); b.writeFloatBE(val); parts.push(b); break; }
      case "float64": { const b = Buffer.alloc(8); b.writeDoubleBE(val); parts.push(b); break; }
      case "bool": parts.push(Buffer.from([val ? 1 : 0])); break;
      case "string": {
        const strBuf = Buffer.from(String(val), "utf8");
        const lenBuf = Buffer.alloc(4); lenBuf.writeUInt32BE(strBuf.length);
        parts.push(lenBuf, strBuf);
        break;
      }
      case "bytes": {
        const bytesBuf = Buffer.isBuffer(val) ? val : Buffer.from(val, "hex");
        const lb = Buffer.alloc(4); lb.writeUInt32BE(bytesBuf.length);
        parts.push(lb, bytesBuf);
        break;
      }
    }
  }
  return Buffer.concat(parts);
}

function decode(schema, buffer) {
  const buf = Buffer.isBuffer(buffer) ? buffer : Buffer.from(buffer);
  const tagMap = {};
  for (const [name, spec] of Object.entries(schema.fields)) {
    tagMap[spec.tag || 0] = { name, type: spec.type };
  }

  const result = {};
  let offset = 0;
  while (offset < buf.length) {
    const tag = buf[offset++];
    const typeId = buf[offset++];
    const field = tagMap[tag];
    if (!field) { offset++; continue; }

    switch (field.type) {
      case "uint8": result[field.name] = buf[offset++]; break;
      case "uint16": result[field.name] = buf.readUInt16BE(offset); offset += 2; break;
      case "uint32": result[field.name] = buf.readUInt32BE(offset); offset += 4; break;
      case "int8": result[field.name] = buf.readInt8(offset); offset += 1; break;
      case "int16": result[field.name] = buf.readInt16BE(offset); offset += 2; break;
      case "int32": result[field.name] = buf.readInt32BE(offset); offset += 4; break;
      case "float32": result[field.name] = buf.readFloatBE(offset); offset += 4; break;
      case "float64": result[field.name] = buf.readDoubleBE(offset); offset += 8; break;
      case "bool": result[field.name] = buf[offset++] === 1; break;
      case "string": {
        const len = buf.readUInt32BE(offset); offset += 4;
        result[field.name] = buf.slice(offset, offset + len).toString("utf8"); offset += len;
        break;
      }
      case "bytes": {
        const len = buf.readUInt32BE(offset); offset += 4;
        result[field.name] = buf.slice(offset, offset + len); offset += len;
        break;
      }
    }
  }
  return result;
}
`

// GetSerializationRuntime returns the serialization runtime string.
func GetSerializationRuntime() string {
	return SerializationRuntime
}
