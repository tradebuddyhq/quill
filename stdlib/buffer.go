package stdlib

// BufferRuntime provides built-in binary/buffer support.
const BufferRuntime = `
// Quill Buffer Runtime

function toBuffer(data, encoding) {
  if (Buffer.isBuffer(data)) return data;
  return Buffer.from(data, encoding || "utf8");
}

function fromBuffer(buf, encoding) {
  return buf.toString(encoding || "utf8");
}

function toBase64(data) {
  return Buffer.from(data).toString("base64");
}

function fromBase64(str) {
  return Buffer.from(str, "base64").toString("utf8");
}

function toHex(data) {
  return Buffer.from(data).toString("hex");
}

function fromHex(str) {
  return Buffer.from(str, "hex").toString("utf8");
}

function concatBuffers(buffers) {
  return Buffer.concat(buffers.map(b => Buffer.isBuffer(b) ? b : Buffer.from(b)));
}
`

// GetBufferRuntime returns the buffer runtime string.
func GetBufferRuntime() string {
	return BufferRuntime
}
