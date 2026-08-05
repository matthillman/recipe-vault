import { handleImport } from "../_lib/import-core.js";

export function onRequestPost(context) {
  return handleImport(context.request, context.env);
}

export function onRequestOptions() {
  return new Response(null, {
    status: 204,
    headers: {
      Allow: "POST, OPTIONS",
      "Access-Control-Allow-Headers": "Authorization, Content-Type",
      "Access-Control-Allow-Methods": "POST, OPTIONS",
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Max-Age": "86400",
    },
  });
}

export function onRequest() {
  return new Response('{"error":"method_not_allowed"}\n', {
    status: 405,
    headers: { "Content-Type": "application/json; charset=utf-8", Allow: "POST, OPTIONS" },
  });
}
