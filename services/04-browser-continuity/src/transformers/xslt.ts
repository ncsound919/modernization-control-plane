/**
 * XSLT Transformer stub — in production this delegates to the Rust/WASM
 * XSLT processor compiled from the rust-core module. For the TypeScript
 * service layer it performs lightweight validation and returns a realistic
 * mock transformation result while emitting structured processing logs.
 */

import {
  XSLTTransformRequest,
  XSLTTransformResponse,
  XSLTValidateRequest,
  XSLTValidateResponse,
} from "../models/index.js";

const XSLT_VERSION_RE = /version\s*=\s*["'](1\.0|2\.0)["']/i;
const XML_DECLARATION_RE = /^<\?xml/;

function detectXsltVersion(xslt: string): "1.0" | "2.0" | "unknown" {
  const match = xslt.match(XSLT_VERSION_RE);
  if (!match) return "unknown";
  return match[1] === "2.0" ? "2.0" : "1.0";
}

function extractRootElement(xml: string): string {
  const match = xml.match(/<([a-zA-Z][a-zA-Z0-9_:-]*)/);
  return match ? match[1] : "root";
}

function buildTransformedOutput(
  xml: string,
  xslt: string,
  params: Record<string, string>
): string {
  const rootEl = extractRootElement(xml);
  const version = detectXsltVersion(xslt);
  const paramEntries = Object.entries(params)
    .map(([k, v]) => `  <!-- param: ${k} = ${v} -->`)
    .join("\n");

  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    `<!-- Transformed by @mcp/browser-continuity XSLT engine (WASM stub) -->`,
    `<!-- Source root element: <${rootEl}> | XSLT version: ${version} -->`,
    `<result>`,
    paramEntries ? `${paramEntries}` : undefined,
    `  <status>transformed</status>`,
    `  <engine>wasm-xslt-stub-v0.1.0</engine>`,
    `  <sourceElement>${rootEl}</sourceElement>`,
    `</result>`,
  ]
    .filter(Boolean)
    .join("\n");
}

export function transformXslt(req: XSLTTransformRequest): XSLTTransformResponse {
  const start = Date.now();
  const warnings: string[] = [];

  console.log("[xslt] transform request received");

  if (!XML_DECLARATION_RE.test(req.xml.trim())) {
    warnings.push("Input XML is missing an <?xml?> declaration — assuming UTF-8.");
  }

  const version = detectXsltVersion(req.xslt);
  if (version === "unknown") {
    warnings.push(
      "Could not detect XSLT version from stylesheet; defaulting to 1.0 semantics."
    );
  }

  if (version === "2.0") {
    warnings.push(
      "XSLT 2.0 support is experimental in the WASM core; some features may fall back to 1.0 behaviour."
    );
  }

  console.log(`[xslt] detected version=${version}, params=${JSON.stringify(req.params ?? {})}`);

  const output = buildTransformedOutput(req.xml, req.xslt, req.params ?? {});
  const processingTimeMs = Date.now() - start;

  console.log(`[xslt] transform complete in ${processingTimeMs}ms, warnings=${warnings.length}`);

  return { output, warnings, processingTimeMs };
}

export function validateXslt(req: XSLTValidateRequest): XSLTValidateResponse {
  const errors: string[] = [];
  const warnings: string[] = [];

  console.log("[xslt] validate request received");

  const version = detectXsltVersion(req.xslt);

  if (!req.xslt.includes("<xsl:stylesheet") && !req.xslt.includes("<xsl:transform")) {
    errors.push("Stylesheet must have a root <xsl:stylesheet> or <xsl:transform> element.");
  }

  if (!req.xslt.includes('xmlns:xsl="http://www.w3.org/1999/XSL/Transform"')) {
    warnings.push("XSLT namespace declaration (xmlns:xsl) not detected; stylesheet may be invalid.");
  }

  if (version === "2.0") {
    warnings.push(
      "XSLT 2.0 stylesheets require the WASM 2.0 engine tier; ensure WASM_XSLT_VERSION=2 is set."
    );
  }

  const valid = errors.length === 0;
  console.log(`[xslt] validate complete: valid=${valid}, errors=${errors.length}`);

  return { valid, version, errors, warnings };
}
