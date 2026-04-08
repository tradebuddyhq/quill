package stdlib

// VectorRuntime provides embedding and vector search support.
const VectorRuntime = `
// Vector / Embedding runtime
async function embed(text, provider, model) {
    provider = provider || "openai";
    if (provider === "openai") {
        const __openai_lib = require("openai");
        const client = new __openai_lib.default();
        const resp = await client.embeddings.create({
            model: model || "text-embedding-3-small",
            input: text
        });
        return resp.data[0].embedding;
    } else if (provider === "ollama") {
        const resp = await fetch("http://localhost:11434/api/embeddings", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ model: model || "nomic-embed-text", prompt: text })
        });
        const data = await resp.json();
        return data.embedding;
    }
    throw new Error("Unsupported embedding provider: " + provider);
}

function cosineSimilarity(a, b) {
    let dot = 0, magA = 0, magB = 0;
    for (let i = 0; i < a.length; i++) {
        dot += a[i] * b[i];
        magA += a[i] * a[i];
        magB += b[i] * b[i];
    }
    return dot / (Math.sqrt(magA) * Math.sqrt(magB));
}

class VectorStore {
    constructor() { this.items = []; }

    async add(text, metadata, provider, model) {
        const vector = await embed(text, provider, model);
        this.items.push({ text, vector, metadata: metadata || {} });
    }

    async addMany(texts, metadatas, provider, model) {
        for (let i = 0; i < texts.length; i++) {
            await this.add(texts[i], metadatas ? metadatas[i] : {}, provider, model);
        }
    }

    async search(query, topK, provider, model) {
        topK = topK || 5;
        const queryVector = await embed(query, provider, model);
        const scored = this.items.map(item => ({
            text: item.text,
            metadata: item.metadata,
            score: cosineSimilarity(queryVector, item.vector)
        }));
        scored.sort((a, b) => b.score - a.score);
        return scored.slice(0, topK);
    }

    toJSON() {
        return JSON.stringify(this.items);
    }

    static fromJSON(json) {
        const store = new VectorStore();
        store.items = JSON.parse(json);
        return store;
    }
}

function createVectorStore() {
    return new VectorStore();
}
`

// GetVectorRuntime returns the vector runtime string.
func GetVectorRuntime() string {
	return VectorRuntime
}
