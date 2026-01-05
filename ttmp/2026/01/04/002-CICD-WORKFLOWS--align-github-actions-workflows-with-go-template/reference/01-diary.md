---
Title: Diary
Ticket: 002-CICD-WORKFLOWS
Status: active
Topics:
    - ci
    - cicd
    - github-actions
    - go
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: .github/workflows/dependency-scanning.yml
      Note: Workflow changes described in Step 1
    - Path: .github/workflows/release.yaml
      Note: Workflow changes described in Step 1
    - Path: .github/workflows/release.yml
      Note: Workflow changes described in Step 1
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-04T19:21:25.758263818-05:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Track changes made to align this repo’s GitHub Actions workflows with the `go-template` baseline, including rationale and validation commands.

## Step 1: Sync GitHub Actions workflows with go-template

This step updates the repo’s CI/security/release workflow definitions to match the canonical versions from `/home/manuel/code/wesen/corporate-headquarters/go-template/.github/workflows`. The intent is to keep workflows consistent across Go repos and reduce drift in scanning and release automation.

This is a CI-only change (YAML updates); runtime and CLI behavior are unchanged.

**Commit (code):** a274c2c — "CI: sync workflows with go-template"

### What I did
- Synced workflow files to match go-template:
  - `.github/workflows/dependency-scanning.yml`
  - `.github/workflows/release.yml`
  - `.github/workflows/release.yaml`
- Closed ticket `002-CICD-WORKFLOWS` once all tasks were complete (`docmgr ticket close ...`).
- Ran:
  - `GOWORK=off go test ./...`
  - `GOWORK=off golangci-lint run ./...`
  - `bash test-scripts/test-cli.sh`

### Why
- Keep GitHub Actions workflows consistent across projects so CI behavior (security scanning and releases) is predictable and maintained in one place (go-template).

### What worked
- Workflows now match the template exactly (verified via `diff -u`).
- Local validation commands passed.

### What didn't work
- N/A

### What I learned
- The local pre-push hook failures are workspace `go.work`-driven; the repo itself does not include a `go.work`, so GitHub Actions is unaffected.

### What was tricky to build
- Ensuring the workflow files match the template exactly (including removing an extra scan job and normalizing YAML EOF newlines).

### What warrants a second pair of eyes
- Confirm removing the extra dependency scan job (`nancy`) is desired for this repo and consistent with the template’s security posture.

### What should be done in the future
- If go-template workflows change, rerun this sync and keep a short changelog entry so security/release automation doesn’t drift.

### Code review instructions
- Start with `.github/workflows/dependency-scanning.yml`, then check `.github/workflows/release.yml` and `.github/workflows/release.yaml`.
- Validate with:
  - `GOWORK=off go test ./...`
  - `GOWORK=off golangci-lint run ./...`

<!-- Provide background context needed to use this reference -->

## Quick Reference

<!-- Provide copy/paste-ready content, API contracts, or quick-look tables -->

## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->
