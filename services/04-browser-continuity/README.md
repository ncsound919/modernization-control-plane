# Module 04 — Browser Continuity Engine

## Purpose
Prevents legacy web portals and embedded device UIs from breaking due to browser deprecations. Provides transparent WASM-based polyfills and MV3-compliant extension shims.

## Browser Deprecation Events Handled

| Deprecation | Timeline | Impact |
|---|---|---|
| XSLT / XSLTProcessor removal from Chromium | ~Nov 2026 (v155+) | Legacy portals rendering XML break |
| Manifest V2 extension phase-out | Chrome 138/139 (2026) | Enterprise browser extensions stop working |

## Capabilities

### WASM XSLT Transcoder
- Compiles a sandboxed XSLT 1.0/2.0 processor to WebAssembly
- Intercepts `XSLTProcessor` usage in legacy portals transparently
- Client-side JS SDK: `import { transcodeXmlToHtml } from '@mcp/xslt-polyfill'`
- Zero server-side changes required for existing portals
- Target use cases: healthcare radiology portals, manufacturing HMIs, EDI outputs

### MV3 Extension Shim
- Translates legacy MV2 extension capabilities to MV3-compliant patterns
- In-page JS SDK for capabilities that can't live in service workers
- Enterprise policy deployment via MDM/GPO

## Tech Stack
- **WASM Runtime:** Rust compiled to WASM via `wasm-pack`
- **XSLT Engine:** Saxon-HE port / libxslt via Emscripten
- **JS SDK:** TypeScript, ESM + CJS bundles
- **Extension:** Chrome Extension API v3, TypeScript
- **Testing:** Playwright (cross-browser), Vitest

## Directory Structure
```
04-browser-continuity/
├── xslt-polyfill/
│   ├── rust-core/        # Rust XSLT processor -> WASM
│   └── js-sdk/           # TypeScript wrapper and auto-intercept
├── mv3-shim/
│   ├── extension/        # MV3 Chrome extension
│   └── in-page-sdk/      # In-page capabilities JS SDK
├── tests/                # Playwright e2e tests
└── README.md
```

## Key Metrics
- Number of portals kept functional post-XSLT removal
- % of legacy MV2 workflows migrated to MV3 runtime
- WASM bundle size (target < 500KB gzipped)

## Status
`Phase 3 — Planned`
