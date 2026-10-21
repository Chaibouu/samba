// Minimal Posta client for Node 18+ (uses the built-in fetch, no dependency).
//
//   POSTA_URL=https://posta.example.com POSTA_API_KEY=psk_... node posta.mjs
//
// Copy this file into a project, or install the official SDK instead:
// https://github.com/goposta/posta-node

import { pathToFileURL } from "node:url";

const POSTA_URL = process.env.POSTA_URL;
const POSTA_API_KEY = process.env.POSTA_API_KEY;

async function call(path, body) {
  const res = await fetch(`${POSTA_URL}/api/v1${path}`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${POSTA_API_KEY}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  const json = await res.json().catch(() => ({}));
  if (!res.ok || json.success === false) {
    const msg = json.error?.message ?? res.statusText;
    throw new Error(`Posta ${res.status}: ${msg}`);
  }
  return json.data; // { id, status }
}

/**
 * Send an email.
 * @param {{from: string, to: string[], subject: string, html?: string, text?: string,
 *          headers?: Record<string,string>, send_at?: string}} email
 */
export function sendEmail(email) {
  return call("/emails/send", email);
}

/**
 * Send an email from a template created in the dashboard.
 * @param {{template: string, to: string[], from?: string, language?: string,
 *          template_data?: Record<string, unknown>}} email
 */
export function sendTemplate(email) {
  return call("/emails/send-template", email);
}

// Quick test when run directly: node posta.mjs you@example.com
if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const to = process.argv[2];
  if (!to) {
    console.error("usage: node posta.mjs <recipient>");
    process.exit(1);
  }
  sendEmail({
    from: process.env.POSTA_FROM ?? "Test <hello@example.com>",
    to: [to],
    subject: "Hello from Posta",
    html: "<h1>It works!</h1><p>Sent through Posta.</p>",
  })
    .then((r) => console.log("queued:", r))
    .catch((e) => {
      console.error(e.message);
      process.exit(1);
    });
}
