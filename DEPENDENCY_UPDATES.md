# Dependency Updates Audit

This document summarizes the dependency updates performed in this project.

## Direct Dependencies

The following direct dependencies were audited and found to already be at their latest available stable versions. No updates were required.

- `github.com/charmbracelet/bubbles`: `v1.0.0`
- `github.com/charmbracelet/bubbletea`: `v1.3.10`
- `github.com/charmbracelet/lipgloss`: `v1.1.0`

## Transitive Dependencies

The following transitive dependencies were safely updated using `go get -u ./...` followed by `go mod tidy` to ensure the project benefits from the latest security patches, bug fixes, and performance improvements:

| Dependency | Old Version | New Version | Reason for Update |
|---|---|---|---|
| `github.com/charmbracelet/colorprofile` | `v0.4.1` | `v0.4.3` | Routine safe patch update |
| `github.com/charmbracelet/x/exp/charmtone` | `v0.0.0-20260323091123-df7b1bcffcca` | `v0.0.0-20260330094520-2dce04b6f8a4` | Routine safe patch update |
| `github.com/clipperhouse/displaywidth` | `v0.9.0` | `v0.11.0` | Routine safe minor update |
| `github.com/clipperhouse/uax29/v2` | `v2.5.0` | `v2.7.0` | Routine safe minor update |
| `github.com/lucasb-eyer/go-colorful` | `v1.3.0` | `v1.4.0` | Routine safe minor update |
| `github.com/mattn/go-runewidth` | `v0.0.19` | `v0.0.22` | Routine safe patch update |
| `golang.org/x/sys` | `v0.41.0` | `v0.42.0` | Routine safe minor update |
| `golang.org/x/text` | `v0.31.0` | `v0.35.0` | Routine safe minor update |

## Build Integrity and Testing

Following the updates, the project's build integrity was verified. Compilation succeeded for all targeted platforms (`linux/amd64`, `linux/arm64`, `windows/amd64`, `windows/arm64`, `darwin/amd64`, `darwin/arm64`) using `./build.sh --all`.
