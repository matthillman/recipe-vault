import test from "node:test";
import assert from "node:assert/strict";

import { decodeImportIssue } from "./decode-import-issue.mjs";

function issueFor(payload) {
  const encoded = Buffer.from(JSON.stringify(payload), "utf8").toString("base64");
  return { body: `Queued by recipe capture.\n\n<!-- recipe-import:v1\n${encoded}\n-->` };
}

test("decodes a versioned import payload", () => {
  assert.deepEqual(
    decodeImportIssue(issueFor({ url: "https://example.com/recipe", client: "ios-shortcut" })),
    { url: "https://example.com/recipe", client: "ios-shortcut" },
  );
});

test("rejects unsupported fields and insecure URLs", () => {
  assert.throws(() => decodeImportIssue(issueFor({ url: "http://example.com" })), /must use https/);
  assert.throws(() => decodeImportIssue(issueFor({ sourceText: "recipe", extra: "no" })), /unsupported/);
});
