package stdlib

// AIProvidersRuntime provides multi-provider LLM support.
const AIProvidersRuntime = `
// Multi-provider AI runtime
const __openai_lib = require("openai");
const __openai_client = new __openai_lib.default();

async function __ask_openai(prompt, options = {}) {
    const model = options.model || "gpt-4o";
    const max_tokens = options.max_tokens || 1024;
    const messages = Array.isArray(prompt) ? prompt : [{ role: "user", content: prompt }];
    if (options.system) messages.unshift({ role: "system", content: options.system });
    const resp = await __openai_client.chat.completions.create({
        model, max_tokens, messages,
        ...(options.temperature !== undefined ? { temperature: options.temperature } : {})
    });
    return resp.choices[0].message.content;
}

async function* __stream_openai(prompt, options = {}) {
    const model = options.model || "gpt-4o";
    const max_tokens = options.max_tokens || 1024;
    const messages = Array.isArray(prompt) ? prompt : [{ role: "user", content: prompt }];
    if (options.system) messages.unshift({ role: "system", content: options.system });
    const stream = await __openai_client.chat.completions.create({
        model, max_tokens, messages, stream: true,
        ...(options.temperature !== undefined ? { temperature: options.temperature } : {})
    });
    for await (const chunk of stream) {
        const text = chunk.choices[0]?.delta?.content;
        if (text) yield text;
    }
}

const { GoogleGenerativeAI } = require("@google/generative-ai");
const __gemini_client = new GoogleGenerativeAI(process.env.GEMINI_API_KEY || process.env.GOOGLE_API_KEY);

async function __ask_gemini(prompt, options = {}) {
    const model = __gemini_client.getGenerativeModel({ model: options.model || "gemini-2.0-flash" });
    const promptText = Array.isArray(prompt) ? prompt.map(m => m.content).join("\n") : prompt;
    const result = await model.generateContent(promptText);
    return result.response.text();
}

async function* __stream_gemini(prompt, options = {}) {
    const model = __gemini_client.getGenerativeModel({ model: options.model || "gemini-2.0-flash" });
    const promptText = Array.isArray(prompt) ? prompt.map(m => m.content).join("\n") : prompt;
    const result = await model.generateContentStream(promptText);
    for await (const chunk of result.stream) {
        yield chunk.text();
    }
}

async function __ask_ollama(prompt, options = {}) {
    const model = options.model || "llama3";
    const messages = Array.isArray(prompt) ? prompt : [{ role: "user", content: prompt }];
    if (options.system) messages.unshift({ role: "system", content: options.system });
    const resp = await fetch("http://localhost:11434/api/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ model, messages, stream: false })
    });
    const data = await resp.json();
    return data.message.content;
}

async function* __stream_ollama(prompt, options = {}) {
    const model = options.model || "llama3";
    const messages = Array.isArray(prompt) ? prompt : [{ role: "user", content: prompt }];
    if (options.system) messages.unshift({ role: "system", content: options.system });
    const resp = await fetch("http://localhost:11434/api/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ model, messages, stream: true })
    });
    const reader = resp.body;
    const decoder = new (require("string_decoder").StringDecoder)("utf8");
    for await (const chunk of reader) {
        const lines = decoder.write(chunk).split("\n").filter(l => l.trim());
        for (const line of lines) {
            try {
                const data = JSON.parse(line);
                if (data.message?.content) yield data.message.content;
            } catch (e) {}
        }
    }
}

function __parse_structured(text, schema) {
    // Try to extract JSON from the response
    let jsonStr = text;
    const jsonMatch = text.match(/` + "\\`\\`\\`" + `json\s*([\s\S]*?)` + "\\`\\`\\`" + `/) || text.match(/\{[\s\S]*\}/);
    if (jsonMatch) jsonStr = jsonMatch[1] || jsonMatch[0];
    try {
        const parsed = JSON.parse(jsonStr.trim());
        const result = {};
        for (const [key, type] of Object.entries(schema)) {
            if (key in parsed) {
                switch (type) {
                    case "number": result[key] = Number(parsed[key]); break;
                    case "bool": result[key] = Boolean(parsed[key]); break;
                    case "list": result[key] = Array.isArray(parsed[key]) ? parsed[key] : [parsed[key]]; break;
                    default: result[key] = String(parsed[key]);
                }
            }
        }
        return result;
    } catch (e) {
        return { error: "Failed to parse structured output", raw: text };
    }
}
`

// GetAIProvidersRuntime returns the AI providers runtime string.
func GetAIProvidersRuntime() string {
	return AIProvidersRuntime
}
