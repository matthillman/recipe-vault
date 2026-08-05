#!/usr/bin/env node

import fs from "node:fs/promises";

export function decodeImportIssue(issue) {
  const body = typeof issue?.body === "string" ? issue.body : "";
  const match = /<!-- recipe-import:v1\s+([A-Za-z0-9+/=]+)\s+-->/m.exec(body);
  if (!match) throw new Error("issue does not contain a recipe-import:v1 payload");

  let request;
  try {
    request = JSON.parse(Buffer.from(match[1], "base64").toString("utf8"));
  } catch (error) {
    throw new Error(`invalid recipe import payload: ${error.message}`);
  }
  if (!request || typeof request !== "object" || Array.isArray(request)) {
    throw new Error("recipe import payload must be an object");
  }
  const allowed = ["url", "sourceText", "note", "client", "idempotencyKey"];
  for (const key of Object.keys(request)) {
    if (!allowed.includes(key)) throw new Error(`unsupported request field: ${key}`);
    if (typeof request[key] !== "string") throw new Error(`${key} must be a string`);
  }
  if (!request.url && !request.sourceText) throw new Error("payload must include url and/or sourceText");
  if (request.url) {
    const parsed = new URL(request.url);
    if (parsed.protocol !== "https:") throw new Error("url must use https");
    if (request.url.length > 2048) throw new Error("url exceeds 2048 characters");
  }
  if ((request.sourceText?.length ?? 0) > 32 * 1024) throw new Error("sourceText exceeds 32 KiB");
  if ((request.note?.length ?? 0) > 1024) throw new Error("note exceeds 1 KiB");
  return request;
}

async function main(argv) {
  const issueIndex = argv.indexOf("--issue");
  const outIndex = argv.indexOf("--out");
  if (issueIndex < 0 || !argv[issueIndex + 1] || outIndex < 0 || !argv[outIndex + 1]) {
    throw new Error("usage: decode-import-issue.mjs --issue issue.json --out request.json");
  }
  const issue = JSON.parse(await fs.readFile(argv[issueIndex + 1], "utf8"));
  const request = decodeImportIssue(issue);
  await fs.writeFile(argv[outIndex + 1], `${JSON.stringify(request, null, 2)}\n`, "utf8");
}

if (import.meta.url === `file://${process.argv[1]}`) {
  main(process.argv.slice(2)).catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
  });
}
