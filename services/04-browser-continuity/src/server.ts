import express, { Request, Response, NextFunction } from "express";
import { transformXslt, validateXslt } from "./transformers/xslt.js";
import { analyzeExtension } from "./transformers/mv3shim.js";
import {
  XSLTTransformRequest,
  XSLTValidateRequest,
  ExtensionManifest,
  PolyfillStatus,
  CompatibilityReport,
  ApiResponse,
} from "./models/index.js";

const app = express();
app.use(express.json({ limit: "4mb" }));

const PORT = parseInt(process.env.PORT ?? "8080", 10);
const SERVICE_START = new Date().toISOString();

function ok<T>(data: T): ApiResponse<T> {
  return { success: true, data, timestamp: new Date().toISOString() };
}

// ---------------------------------------------------------------------------
// GET /health
// ---------------------------------------------------------------------------
app.get("/health", (_req: Request, res: Response) => {
  res.json({
    status: "healthy",
    service: "@mcp/browser-continuity",
    version: "0.1.0",
    startedAt: SERVICE_START,
    timestamp: new Date().toISOString(),
    capabilities: ["xslt-transform", "xslt-validate", "mv3-analysis", "polyfills", "compatibility"],
  });
});

// ---------------------------------------------------------------------------
// POST /api/v1/xslt/transform
// ---------------------------------------------------------------------------
app.post("/api/v1/xslt/transform", (req: Request, res: Response) => {
  const body = req.body as Partial<XSLTTransformRequest>;

  if (!body.xml || !body.xslt) {
    res.status(400).json({
      success: false,
      error: "Both 'xml' and 'xslt' fields are required.",
      timestamp: new Date().toISOString(),
    });
    return;
  }

  const result = transformXslt({
    xml: body.xml,
    xslt: body.xslt,
    params: body.params ?? {},
  });

  res.json(ok(result));
});

// ---------------------------------------------------------------------------
// POST /api/v1/xslt/validate
// ---------------------------------------------------------------------------
app.post("/api/v1/xslt/validate", (req: Request, res: Response) => {
  const body = req.body as Partial<XSLTValidateRequest>;

  if (!body.xslt) {
    res.status(400).json({
      success: false,
      error: "'xslt' field is required.",
      timestamp: new Date().toISOString(),
    });
    return;
  }

  const result = validateXslt({ xslt: body.xslt });
  res.status(result.valid ? 200 : 422).json(ok(result));
});

// ---------------------------------------------------------------------------
// GET /api/v1/extensions
// ---------------------------------------------------------------------------
app.get("/api/v1/extensions", (_req: Request, res: Response) => {
  const extensions = [
    {
      id: "ext-001",
      name: "Legacy SSO Bridge",
      manifestVersion: 2,
      migrationStatus: "pending",
      installedVersion: "3.2.1",
      lastChecked: "2025-06-01T00:00:00Z",
    },
    {
      id: "ext-002",
      name: "Enterprise Proxy Helper",
      manifestVersion: 2,
      migrationStatus: "in-progress",
      installedVersion: "1.9.4",
      lastChecked: "2025-07-01T00:00:00Z",
    },
    {
      id: "ext-003",
      name: "MFA Token Injector",
      manifestVersion: 3,
      migrationStatus: "complete",
      installedVersion: "2.0.0",
      lastChecked: "2025-08-01T00:00:00Z",
    },
  ];
  res.json(ok(extensions));
});

// ---------------------------------------------------------------------------
// POST /api/v1/extensions/analyze
// ---------------------------------------------------------------------------
app.post("/api/v1/extensions/analyze", (req: Request, res: Response) => {
  const body = req.body as Partial<{ manifest: ExtensionManifest }>;

  if (!body.manifest || typeof body.manifest !== "object") {
    res.status(400).json({
      success: false,
      error: "'manifest' object is required in the request body.",
      timestamp: new Date().toISOString(),
    });
    return;
  }

  if (typeof body.manifest.manifest_version !== "number") {
    res.status(400).json({
      success: false,
      error: "manifest.manifest_version must be a number.",
      timestamp: new Date().toISOString(),
    });
    return;
  }

  const analysis = analyzeExtension(body.manifest);
  res.json(ok(analysis));
});

// ---------------------------------------------------------------------------
// GET /api/v1/polyfills
// ---------------------------------------------------------------------------
app.get("/api/v1/polyfills", (_req: Request, res: Response) => {
  const polyfills: PolyfillStatus[] = [
    {
      name: "@mcp/xslt-polyfill",
      version: "0.1.0",
      browserSupport: {
        chrome: ">=100",
        firefox: ">=91",
        safari: ">=15",
        edge: ">=100",
      },
      status: "beta",
      description:
        "WASM-backed XSLTProcessor polyfill. Transparently intercepts XSLTProcessor instantiation and delegates to the sandboxed XSLT 1.0/2.0 WASM engine.",
      bundleSizeKb: 312,
    },
    {
      name: "@mcp/mv3-content-shim",
      version: "0.2.1",
      browserSupport: {
        chrome: ">=112",
        firefox: ">=115",
        safari: ">=17",
        edge: ">=112",
      },
      status: "stable",
      description:
        "In-page SDK for MV3-incompatible extension capabilities that must execute in page context rather than a service worker.",
      bundleSizeKb: 48,
    },
    {
      name: "@mcp/declarative-net-request-bridge",
      version: "0.1.2",
      browserSupport: {
        chrome: ">=88",
        firefox: ">=113",
        safari: ">=15.4",
        edge: ">=88",
      },
      status: "experimental",
      description:
        "Converts legacy chrome.webRequest blocking rule sets to declarativeNetRequest rule objects.",
      bundleSizeKb: 22,
    },
  ];
  res.json(ok(polyfills));
});

// ---------------------------------------------------------------------------
// GET /api/v1/compatibility
// ---------------------------------------------------------------------------
app.get("/api/v1/compatibility", (_req: Request, res: Response) => {
  const reports: CompatibilityReport[] = [
    {
      browser: "Chrome",
      version: "138",
      deprecations: [
        {
          api: "XSLTProcessor",
          removedInVersion: "155",
          removalDate: "2026-11-01",
          replacement: null,
          polyfillAvailable: true,
        },
        {
          api: "Manifest V2 extensions",
          removedInVersion: "139",
          removalDate: "2026-06-01",
          replacement: "Manifest V3",
          polyfillAvailable: false,
        },
        {
          api: "chrome.tabs.executeScript",
          removedInVersion: "120",
          removalDate: "2023-11-01",
          replacement: "chrome.scripting.executeScript",
          polyfillAvailable: false,
        },
      ],
      checkedAt: new Date().toISOString(),
    },
    {
      browser: "Firefox",
      version: "128",
      deprecations: [
        {
          api: "Manifest V2 extensions",
          removedInVersion: "140",
          removalDate: "2027-01-01",
          replacement: "Manifest V3",
          polyfillAvailable: false,
        },
      ],
      checkedAt: new Date().toISOString(),
    },
    {
      browser: "Safari",
      version: "17.5",
      deprecations: [],
      checkedAt: new Date().toISOString(),
    },
  ];
  res.json(ok(reports));
});

// ---------------------------------------------------------------------------
// Global error handler
// ---------------------------------------------------------------------------
app.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
  console.error("[server] unhandled error:", err.message);
  res.status(500).json({
    success: false,
    error: "Internal server error.",
    timestamp: new Date().toISOString(),
  });
});

app.listen(PORT, () => {
  console.log(`[server] @mcp/browser-continuity listening on port ${PORT}`);
});

export default app;
