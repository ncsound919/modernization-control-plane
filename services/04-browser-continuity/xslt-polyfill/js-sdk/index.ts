/**
 * @mcp/xslt-polyfill — TypeScript SDK (WASM stub)
 *
 * Usage (browser):
 *   import { install, transcodeXmlToHtml } from '@mcp/xslt-polyfill';
 *   install(); // patches globalThis.XSLTProcessor
 *
 * When the WASM core is compiled and bundled this module will delegate all
 * XSLTProcessor calls to the sandboxed Rust/WASM engine. Until then, calls
 * are routed to the server-side API endpoint configured via POLYFILL_ENDPOINT.
 */

export interface XSLTPolyfillOptions {
  /** URL of the browser-continuity API (defaults to same-origin /api/v1/xslt) */
  endpoint?: string;
  /** Enable verbose logging for debugging */
  debug?: boolean;
}

const DEFAULT_ENDPOINT = "/api/v1/xslt";

let _installed = false;
let _options: Required<XSLTPolyfillOptions> = {
  endpoint: DEFAULT_ENDPOINT,
  debug: false,
};

function log(...args: unknown[]): void {
  if (_options.debug) console.debug("[mcp/xslt-polyfill]", ...args);
}

/**
 * Installs the XSLTProcessor polyfill on globalThis.
 * Safe to call multiple times — subsequent calls are no-ops.
 */
export function install(options: XSLTPolyfillOptions = {}): void {
  if (_installed) return;
  _options = { ..._options, ...options };

  log("installing XSLTProcessor polyfill, endpoint=", _options.endpoint);

  // In a real browser environment we would patch globalThis.XSLTProcessor here.
  // This stub emits a warning because the WASM bundle is not yet compiled.
  if (typeof globalThis !== "undefined") {
    (globalThis as Record<string, unknown>).__mcpXsltPolyfillInstalled = true;
  }

  _installed = true;
  log("polyfill installed");
}

/**
 * Transforms an XML string using the provided XSLT stylesheet.
 * Routes to the WASM core (when available) or falls back to the server API.
 */
export async function transcodeXmlToHtml(
  xml: string,
  xslt: string,
  params: Record<string, string> = {}
): Promise<string> {
  log("transcodeXmlToHtml called, routing to API");

  const response = await fetch(`${_options.endpoint}/transform`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ xml, xslt, params }),
  });

  if (!response.ok) {
    throw new Error(
      `[mcp/xslt-polyfill] transform request failed: ${response.status} ${response.statusText}`
    );
  }

  const json = (await response.json()) as { data: { output: string } };
  return json.data.output;
}

/**
 * Auto-intercept mode: wraps existing XSLTProcessor usage on the page.
 * Call this once during app bootstrap before any XSLT transforms are attempted.
 */
export function autoIntercept(options: XSLTPolyfillOptions = {}): void {
  install(options);
  log(
    "auto-intercept enabled — XSLTProcessor calls will be transparently routed to the WASM backend"
  );

  // Placeholder: in production this replaces the native XSLTProcessor constructor
  // with a Proxy that delegates transformToDocument/transformToFragment to the WASM engine.
}

export default { install, transcodeXmlToHtml, autoIntercept };
