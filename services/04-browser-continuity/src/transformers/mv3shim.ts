/**
 * MV3 Shim Analyzer — parses a Manifest V2 Chrome extension manifest and
 * produces a structured list of migration steps needed to make the extension
 * compatible with Manifest V3.  In production, deeper static analysis of the
 * extension bundle would supplement these manifest-level checks.
 */

import { ExtensionManifest, ExtensionAnalysis, MigrationStep } from "../models/index.js";

const DOCS_BASE = "https://developer.chrome.com/docs/extensions/develop/migrate";

const DEPRECATED_PERMISSION_MAP: Record<string, string> = {
  background: "Service workers do not require a background permission.",
  unlimited_storage: "Use chrome.storage.session or chrome.storage.local with quotas instead.",
};

const DEPRECATED_API_MAP: Record<string, string> = {
  "chrome.browserAction": "chrome.action",
  "chrome.pageAction": "chrome.action",
  "chrome.webRequest (blocking)":
    "chrome.declarativeNetRequest for blocking; webRequest for observation only",
  "chrome.extension.getBackgroundPage": "chrome.runtime.getBackgroundPage (deprecated in MV3)",
  "chrome.tabs.executeScript": "chrome.scripting.executeScript",
  "chrome.tabs.insertCSS": "chrome.scripting.insertCSS",
};

function detectBackgroundType(
  manifest: ExtensionManifest
): ExtensionAnalysis["backgroundType"] {
  if (!manifest.background) return "none";
  if (manifest.background.service_worker) return "service-worker";
  if (manifest.background.persistent === false) return "event-page";
  if (manifest.background.scripts || manifest.background.page)
    return "persistent-background-page";
  return "none";
}

function buildMigrationSteps(manifest: ExtensionManifest): MigrationStep[] {
  const steps: MigrationStep[] = [];
  const bgType = detectBackgroundType(manifest);

  // Manifest version bump
  steps.push({
    area: "manifest",
    severity: "breaking",
    description: "Update manifest_version from 2 to 3.",
    before: '"manifest_version": 2',
    after: '"manifest_version": 3',
    docsUrl: `${DOCS_BASE}/checklist`,
  });

  // Background page → service worker
  if (bgType === "persistent-background-page" || bgType === "event-page") {
    const bgBefore =
      bgType === "persistent-background-page"
        ? '"background": { "scripts": ["background.js"], "persistent": true }'
        : '"background": { "scripts": ["background.js"], "persistent": false }';

    steps.push({
      area: "background",
      severity: "breaking",
      description:
        "Replace background page / event page with a service worker. Service workers cannot access the DOM and are terminated when idle.",
      before: bgBefore,
      after: '"background": { "service_worker": "background.js" }',
      docsUrl: `${DOCS_BASE}/service-workers`,
    });

    steps.push({
      area: "background",
      severity: "warning",
      description:
        "Service workers are stateless — replace in-memory global variables with chrome.storage.session or chrome.storage.local.",
      before: "let cache = {};  // global, lives forever in background page",
      after: "await chrome.storage.session.set({ cache: {} });",
      docsUrl: `${DOCS_BASE}/service-workers/convert-to-module`,
    });
  }

  // browser_action / page_action → action
  if (manifest.browser_action || manifest.page_action) {
    const oldKey = manifest.browser_action ? "browser_action" : "page_action";
    steps.push({
      area: "action",
      severity: "breaking",
      description: `Replace "${oldKey}" with the unified "action" key introduced in MV3.`,
      before: `"${oldKey}": { "default_popup": "popup.html" }`,
      after: '"action": { "default_popup": "popup.html" }',
      docsUrl: `${DOCS_BASE}/action-api`,
    });
  }

  // web_accessible_resources
  if (
    manifest.web_accessible_resources &&
    Array.isArray(manifest.web_accessible_resources) &&
    typeof manifest.web_accessible_resources[0] === "string"
  ) {
    steps.push({
      area: "web_accessible_resources",
      severity: "breaking",
      description:
        'MV3 requires "web_accessible_resources" to be an array of objects with "resources" and "matches" keys.',
      before: '"web_accessible_resources": ["images/*.png"]',
      after:
        '"web_accessible_resources": [{ "resources": ["images/*.png"], "matches": ["<all_urls>"] }]',
      docsUrl: `${DOCS_BASE}/web-accessible-resources`,
    });
  }

  // Scripting API
  steps.push({
    area: "scripting",
    severity: "breaking",
    description:
      "Replace chrome.tabs.executeScript and chrome.tabs.insertCSS with chrome.scripting equivalents.",
    before: 'chrome.tabs.executeScript(tabId, { file: "content.js" })',
    after: 'chrome.scripting.executeScript({ target: { tabId }, files: ["content.js"] })',
    docsUrl: `${DOCS_BASE}/scripting-api`,
  });

  // Content Security Policy
  steps.push({
    area: "content_security_policy",
    severity: "warning",
    description:
      'MV3 requires "content_security_policy" to be an object with "extension_pages" and optionally "sandbox" keys.',
    before: '"content_security_policy": "script-src \'self\'; object-src \'self\'"',
    after:
      '"content_security_policy": { "extension_pages": "script-src \'self\'; object-src \'self\'" }',
    docsUrl: `${DOCS_BASE}/content-security-policy`,
  });

  // Blocking webRequest
  const perms = manifest.permissions ?? [];
  if (perms.includes("webRequest") && perms.includes("webRequestBlocking")) {
    steps.push({
      area: "network",
      severity: "breaking",
      description:
        "Blocking webRequest is removed in MV3. Migrate to declarativeNetRequest for ad blocking, request modification, and header manipulation.",
      before:
        'chrome.webRequest.onBeforeRequest.addListener(callback, filter, ["blocking"])',
      after:
        'chrome.declarativeNetRequest.updateDynamicRules({ addRules: [...], removeRuleIds: [...] })',
      docsUrl: `${DOCS_BASE}/declarative-net-request`,
    });
  }

  // Permissions that changed
  for (const perm of perms) {
    if (DEPRECATED_PERMISSION_MAP[perm]) {
      steps.push({
        area: "permissions",
        severity: "warning",
        description: DEPRECATED_PERMISSION_MAP[perm],
        before: `"permissions": ["${perm}"]`,
        after: `// Remove "${perm}" from permissions array`,
        docsUrl: `${DOCS_BASE}/checklist#update-permissions`,
      });
    }
  }

  return steps;
}

function detectDeprecatedApis(manifest: ExtensionManifest): string[] {
  const deprecated: string[] = [];
  if (manifest.browser_action) deprecated.push("chrome.browserAction");
  if (manifest.page_action) deprecated.push("chrome.pageAction");

  const perms = manifest.permissions ?? [];
  if (perms.includes("webRequest") && perms.includes("webRequestBlocking")) {
    deprecated.push("chrome.webRequest (blocking)");
  }

  // Always flag scripting and DOM access patterns present in MV2 backgrounds
  deprecated.push("chrome.tabs.executeScript");
  if (detectBackgroundType(manifest) === "persistent-background-page") {
    deprecated.push("chrome.extension.getBackgroundPage");
  }

  return deprecated;
}

function estimateEffort(steps: MigrationStep[]): number {
  const hours = steps.reduce((acc, step) => {
    if (step.severity === "breaking") return acc + 4;
    if (step.severity === "warning") return acc + 1.5;
    return acc + 0.5;
  }, 0);
  return Math.round(hours);
}

export function analyzeExtension(manifest: ExtensionManifest): ExtensionAnalysis {
  console.log(`[mv3shim] analyzing extension: "${manifest.name ?? "unnamed"}" mv${manifest.manifest_version}`);

  const backgroundType = detectBackgroundType(manifest);
  const deprecatedApis = detectDeprecatedApis(manifest);
  const migrationSteps = buildMigrationSteps(manifest);
  const estimatedEffortHours = estimateEffort(migrationSteps);

  console.log(
    `[mv3shim] analysis complete: steps=${migrationSteps.length}, effort=${estimatedEffortHours}h, bgType=${backgroundType}`
  );

  return {
    extensionName: manifest.name ?? "Unknown Extension",
    manifestVersion: manifest.manifest_version,
    permissions: manifest.permissions ?? [],
    backgroundType,
    deprecatedApis,
    migrationSteps,
    estimatedEffortHours,
  };
}
