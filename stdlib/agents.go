package stdlib

// AgentRuntime provides AI agent loop support.
const AgentRuntime = `
// Agent runtime
class Agent {
    constructor(name, options = {}) {
        this.name = name;
        this.tools = {};
        this.memory = [];
        this.provider = options.provider || "claude";
        this.model = options.model || null;
        this.system = options.system || "You are a helpful AI agent named " + name + ". You have access to tools. Use them when needed. Respond with JSON: {\"tool\": \"toolName\", \"args\": {...}} to call a tool, or {\"done\": true, \"result\": \"...\"} when finished.";
        this.maxTurns = options.maxTurns || 10;
    }

    addTool(name, description, fn) {
        this.tools[name] = { description, fn };
    }

    async run(goal) {
        this.memory.push({ role: "user", content: goal });
        const toolDescriptions = Object.entries(this.tools)
            .map(([name, t]) => name + ": " + t.description)
            .join("\\n");
        const systemPrompt = this.system + "\\n\\nAvailable tools:\\n" + toolDescriptions;

        for (let turn = 0; turn < this.maxTurns; turn++) {
            let response;
            const askOpts = { system: systemPrompt };
            if (this.model) askOpts.model = this.model;

            switch (this.provider) {
                case "openai": response = await __ask_openai(this.memory, askOpts); break;
                case "gemini": response = await __ask_gemini(this.memory, askOpts); break;
                case "ollama": response = await __ask_ollama(this.memory, askOpts); break;
                default: {
                    const __anthropic = require("@anthropic-ai/sdk");
                    const __client = new __anthropic.default();
                    const model = askOpts.model || "claude-sonnet-4-20250514";
                    const resp = await __client.messages.create({
                        model, max_tokens: 4096, system: systemPrompt,
                        messages: this.memory
                    });
                    response = resp.content[0].text;
                }
            }

            this.memory.push({ role: "assistant", content: response });

            try {
                const parsed = JSON.parse(response);
                if (parsed.done) return parsed.result;
                if (parsed.tool && this.tools[parsed.tool]) {
                    const toolResult = await this.tools[parsed.tool].fn(parsed.args || {});
                    const resultStr = typeof toolResult === "string" ? toolResult : JSON.stringify(toolResult);
                    this.memory.push({ role: "user", content: "Tool " + parsed.tool + " returned: " + resultStr });
                }
            } catch (e) {
                return response;
            }
        }
        return "Agent reached max turns without completing.";
    }
}

function createAgent(name, options) {
    return new Agent(name, options);
}
`

// GetAgentRuntime returns the agent runtime string.
func GetAgentRuntime() string {
	return AgentRuntime
}
