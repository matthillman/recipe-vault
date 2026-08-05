export function onRequestGet(context) {
  const origin = new URL(context.request.url).origin;
  const document = {
    openapi: "3.1.0",
    info: {
      title: "Recipe Vault Capture API",
      version: "1.0.0",
      description: "Queue a recipe URL or supplied recipe text for conservative normalization and draft-PR review.",
    },
    servers: [{ url: origin }],
    paths: {
      "/api/import": {
        post: {
          operationId: "captureRecipe",
          summary: "Queue a recipe for import",
          security: [{ bearerAuth: [] }],
          requestBody: {
            required: true,
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  additionalProperties: false,
                  properties: {
                    url: { type: "string", format: "uri", maxLength: 2048, description: "Public HTTPS recipe page." },
                    sourceText: { type: "string", maxLength: 32768, description: "Pasted recipe text when a page cannot be fetched." },
                    note: { type: "string", maxLength: 1024 },
                    client: { type: "string", maxLength: 64, default: "chatgpt-action" },
                    idempotencyKey: { type: "string", maxLength: 128 },
                  },
                  anyOf: [{ required: ["url"] }, { required: ["sourceText"] }],
                },
              },
            },
          },
          responses: {
            202: { description: "Queued as a new GitHub issue." },
            200: { description: "An equivalent open import is already queued." },
            400: { description: "Invalid capture request." },
            401: { description: "Missing or invalid bearer token." },
          },
        },
      },
    },
    components: {
      securitySchemes: {
        bearerAuth: { type: "http", scheme: "bearer" },
      },
    },
  };
  return new Response(`${JSON.stringify(document, null, 2)}\n`, {
    headers: { "Content-Type": "application/json; charset=utf-8", "Cache-Control": "public, max-age=300" },
  });
}
