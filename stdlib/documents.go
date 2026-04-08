package stdlib

// DocumentRuntime provides document processing support.
const DocumentRuntime = `
// Document processing runtime
function chunk(text, size, overlap) {
    size = size || 500;
    overlap = overlap || 50;
    const chunks = [];
    let start = 0;
    while (start < text.length) {
        let end = Math.min(start + size, text.length);
        // Try to break at a sentence boundary
        if (end < text.length) {
            const lastPeriod = text.lastIndexOf(".", end);
            const lastNewline = text.lastIndexOf("\\n", end);
            const breakPoint = Math.max(lastPeriod, lastNewline);
            if (breakPoint > start + size * 0.5) end = breakPoint + 1;
        }
        chunks.push(text.slice(start, end).trim());
        start = end - overlap;
    }
    return chunks.filter(c => c.length > 0);
}

async function extract(filePath) {
    const fs = require("fs");
    const path = require("path");
    const ext = path.extname(filePath).toLowerCase();

    if (ext === ".txt" || ext === ".md" || ext === ".csv") {
        return fs.readFileSync(filePath, "utf8");
    }
    if (ext === ".json") {
        return JSON.stringify(JSON.parse(fs.readFileSync(filePath, "utf8")), null, 2);
    }
    if (ext === ".html" || ext === ".htm") {
        const html = fs.readFileSync(filePath, "utf8");
        return html.replace(/<[^>]*>/g, " ").replace(/\\s+/g, " ").trim();
    }
    if (ext === ".pdf") {
        try {
            const pdfParse = require("pdf-parse");
            const buffer = fs.readFileSync(filePath);
            const data = await pdfParse(buffer);
            return data.text;
        } catch (e) {
            throw new Error("PDF parsing requires 'pdf-parse' package. Run: npm install pdf-parse");
        }
    }
    // Fallback: try reading as text
    return fs.readFileSync(filePath, "utf8");
}

function splitSentences(text) {
    return text.match(/[^.!?]+[.!?]+/g) || [text];
}

function splitParagraphs(text) {
    return text.split(/\\n\\s*\\n/).filter(p => p.trim().length > 0);
}
`

// GetDocumentRuntime returns the document runtime string.
func GetDocumentRuntime() string {
	return DocumentRuntime
}
