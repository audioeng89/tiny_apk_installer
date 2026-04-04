## 2025-02-17 - [Path Traversal via XAPK Manifest]
**Vulnerability:** Found a potential path traversal arbitrary file write vulnerability in `.xapk` installation. The `package_name` field from the user-provided `manifest.json` within an `.xapk` was unvalidated and used directly to build an ADB push destination path (`/sdcard/Android/obb/` + `package_name`).
**Learning:** External archive manifests parsed as JSON can harbor malicious input directly leading to file traversal if used for constructing file paths.
**Prevention:** Always strictly validate the `package_name` (and other extracted metadata used in file paths) using regex against accepted schema standards (e.g., `^[a-zA-Z0-9_]+(\.[a-zA-Z0-9_]+)+$`).
