package stdlib

const DiscordRuntime = `
// Quill Discord Bot Runtime Helpers
// Provides convenience functions for building Discord bots with discord.js

const createBot = (token, intents) => {
  const { Client, GatewayIntentBits } = require("discord.js");
  const intentFlags = intents || [
    GatewayIntentBits.Guilds,
    GatewayIntentBits.GuildMessages,
    GatewayIntentBits.MessageContent
  ];
  const client = new Client({ intents: intentFlags });
  if (token) {
    client.login(token);
  }
  return client;
};
`
