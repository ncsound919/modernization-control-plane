export interface XSLTTransformRequest {
  xml: string;
  xslt: string;
  params?: Record<string, string>;
}

export interface XSLTTransformResponse {
  output: string;
  warnings: string[];
  processingTimeMs: number;
}

export interface XSLTValidateRequest {
  xslt: string;
}

export interface XSLTValidateResponse {
  valid: boolean;
  version: "1.0" | "2.0" | "unknown";
  errors: string[];
  warnings: string[];
}

export interface ExtensionManifest {
  name?: string;
  version?: string;
  manifest_version: number;
  permissions?: string[];
  background?: {
    scripts?: string[];
    page?: string;
    service_worker?: string;
    persistent?: boolean;
  };
  browser_action?: Record<string, unknown>;
  page_action?: Record<string, unknown>;
  content_scripts?: Array<{
    matches: string[];
    js?: string[];
    css?: string[];
    run_at?: string;
  }>;
  web_accessible_resources?: string[] | Array<{ resources: string[]; matches: string[] }>;
}

export interface MigrationStep {
  area: string;
  severity: "breaking" | "warning" | "info";
  description: string;
  before: string;
  after: string;
  docsUrl: string;
}

export interface ExtensionAnalysis {
  extensionName: string;
  manifestVersion: number;
  permissions: string[];
  backgroundType: "persistent-background-page" | "event-page" | "service-worker" | "none";
  deprecatedApis: string[];
  migrationSteps: MigrationStep[];
  estimatedEffortHours: number;
}

export interface BrowserSupport {
  chrome: string;
  firefox: string;
  safari: string;
  edge: string;
}

export interface PolyfillStatus {
  name: string;
  version: string;
  browserSupport: BrowserSupport;
  status: "stable" | "beta" | "experimental";
  description: string;
  bundleSizeKb: number;
}

export interface Deprecation {
  api: string;
  removedInVersion: string;
  removalDate: string;
  replacement: string | null;
  polyfillAvailable: boolean;
}

export interface CompatibilityReport {
  browser: string;
  version: string;
  deprecations: Deprecation[];
  checkedAt: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  timestamp: string;
}
