# Security Policy

## Sovereign & Offline-First Security Architecture

**IFM Chat** is engineered with enterprise-grade privacy and data sovereignty as fundamental core tenets. The application operates strictly as a local client workstation connecting directly to specified AI endpoints without intermediary telemetry servers.

---

## 1. Threat Model & Security Posture

### A. Credential Protection (Zero Plaintext Secrets)
- **Windows Credential Manager Integration (`wincred`)**: Sensitive secrets including the **IFM API Keys** (`IFMChat/APIKey`, `IFMChat/Key_<ID>`) and **Tavily Search API Keys** (`IFMChat/TavilyKey`) are encrypted and managed directly via the operating system's native Windows Vault.
- API keys are **never** stored in SQLite database files, local configuration JSONs, or environment variable dumps.

### B. Local Data Sovereignty & Zero Cloud Leakage
- Conversation records, prompt history, user presets, and metadata are stored exclusively on the host system in `%APPDATA%\IFMChat\database\ifm-chat.db`.
- No analytic telemetry, diagnostic pings, or background usage tracking is embedded within the codebase.

### C. Workspace & Folder Ingestion Safety
- The local codebase scanner operates with read-only access to explicitly user-selected folders.
- Files matching standard ignore patterns (`.git`, `node_modules`, `vendor`, `.env`, secret keys) are omitted from context ingestion to avoid leaking confidential local environment variables to model prompts.

### D. Client-Side Rendering Defense (XSS Prevention)
- All streaming markdown and code blocks are sanitized using a multi-pass pipeline:
  1. `markdown-it` parsing with HTML escaping enabled.
  2. `DOMPurify` sanitizer stripping malicious scripts, iframes, and unsanctioned attributes.
  3. `Shiki` syntax highlighting isolated within inert container nodes.

---

## 2. Supported Versions

Security updates and vulnerability patches are actively maintained for the following releases:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

---

## 3. Reporting a Vulnerability

We take the security of this software and user data seriously. If you discover a security vulnerability, please follow the coordinated disclosure process:

1. **Do not disclose the vulnerability publicly** in GitHub issues, pull requests, or public forums.
2. Email a detailed vulnerability report to:
   - **Email**: `65205840+yogamahhendra@users.noreply.github.com`
3. Please include:
   - A description of the issue and potential security impact.
   - Step-by-step reproduction instructions or a minimal proof-of-concept (PoC).
   - Any remediation suggestions if available.

### Response Timeline
- **Initial Acknowledgment**: Within 48 hours.
- **Triage & Status Update**: Within 5 business days.
- **Remediation Release**: A security release will be compiled and tagged once verified.
