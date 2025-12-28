---
Title: 'Prompt management screens: spec vs implementation'
Ticket: 011-TUI-UPDATE
Status: active
Topics:
    - tui
    - bubbletea
    - prescribe
    - ux
    - cli
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: prescribe/internal/controller/controller.go
      Note: SetPrompt()
    - Path: prescribe/internal/domain/domain.go
      Note: PromptPreset
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T20:00:00-05:00
WhatFor: ""
WhenToUse: ""
---


# Prompt Management Screens: Spec vs Implementation

This document analyzes the prompt management screens from the original TUI spec (`claude-session-TUI-simulation.md`) against the current implementation.

## Screens Analyzed

9. **Edit Prompt Template Screen** (Spec §9)
10. **Select Prompt Preset Screen** (Spec §10)
11. **Save Prompt Preset Dialog** (Spec §11)
12. **External Editor Screen** (Spec §12)

---

## 9. Edit Prompt Template Screen

### Spec Requirements

**Layout:**
- Title: "EDIT PROMPT TEMPLATE"
- Current preset name with `[SESSION]` / `[PRESET]` badge
- **PROMPT TEXT** section: Multi-line textarea showing current template
- **OPTIONS** section: `[P] Load Preset [X] Open in $EDITOR [W] Save as Preset [R] Reset to Default`
- Actions: `Type to edit Ctrl+S Save & Use Esc Cancel`

**Data:**
- Current prompt text
- Current preset ID (if loaded from preset)
- Session flag (true if not saved as preset)

### Current Implementation

**What Exists:**
- ✅ Domain: `PRData.CurrentPrompt` (string) and `CurrentPreset` (*PromptPreset)
- ✅ Controller: `SetPrompt(prompt, preset)`, `LoadPromptPreset(presetID)`
- ✅ Built-in presets: `domain.GetBuiltinPresets()` returns default/detailed/concise/conventional
- ✅ Preset loading: Project/global preset directories supported

**What's Missing:**
- ❌ **Screen**: No Edit Prompt Template screen in TUI
- ❌ **Prompt display**: No UI showing current prompt text
- ❌ **Text editing**: No textarea/input for editing prompt
- ❌ **Preset loading**: No `P` key to load preset
- ❌ **External editor**: No `X` key to open in $EDITOR
- ❌ **Save as preset**: No `W` key to save prompt as preset
- ❌ **Reset**: No `R` key to reset to default

**CLI Equivalents:**
- `prescribe generate --prompt "..."` (set prompt)
- `prescribe generate --preset <id>` (load preset)
- Prompt is stored in session YAML but not displayed/edited in TUI

**Code References:**
- Domain: `internal/domain/domain.go::PRData` (CurrentPrompt, CurrentPreset)
- Controller: `internal/controller/controller.go::SetPrompt()`, `LoadPromptPreset()`
- Built-ins: `internal/domain/domain.go::GetBuiltinPresets()`
- Preset loading: `internal/controller/controller.go::LoadProjectPresets()`, `LoadGlobalPresets()`

---

## 10. Select Prompt Preset Screen

### Spec Requirements

**Layout:**
- Title: "SELECT PROMPT PRESET"
- **BUILT-IN PRESETS** section: List with `[BUILTIN]` badge
- **PROJECT PRESETS** section: List with `[PROJECT]` badge and file path
- **GLOBAL PRESETS** section: List with `[GLOBAL]` badge and file path
- Each preset shows: name, description/preview, location badge
- Actions: `↑↓ Navigate Enter Select V View Full E Edit D Delete Esc Cancel`

**Data:**
- Built-in presets (id, name, preview, fullText)
- Project presets (id, name, description, filePath, fullText)
- Global presets (id, name, description, filePath, fullText)
- Selected index and category

### Current Implementation

**What Exists:**
- ✅ Built-in presets: `domain.GetBuiltinPresets()` (default, detailed, concise, conventional)
- ✅ Project presets: `Controller.LoadProjectPresets()` reads `.pr-builder/prompts/*.yaml`
- ✅ Global presets: `Controller.LoadGlobalPresets()` reads `~/.pr-builder/prompts/*.yaml`
- ✅ Preset structure: `PromptPreset` with ID, Name, Description, Template, Location

**What's Missing:**
- ❌ **Screen**: No Select Prompt Preset screen in TUI
- ❌ **Preset list**: No UI to browse/select presets
- ❌ **Location badges**: No display of `[BUILTIN]` / `[PROJECT]` / `[GLOBAL]`
- ❌ **Preview**: No preview of preset text
- ❌ **View full**: No `V` key to view complete preset text
- ❌ **Edit preset**: No `E` key to edit preset
- ❌ **Delete preset**: No `D` key to delete preset

**CLI Equivalents:**
- `prescribe generate --preset <id>` (load preset by ID)
- No CLI command to list presets (would need to be added)

**Code References:**
- Domain: `internal/domain/domain.go::PromptPreset`, `GetBuiltinPresets()`
- Controller: `internal/controller/controller.go::LoadProjectPresets()`, `LoadGlobalPresets()`, `LoadPromptPreset()`
- Preset YAML: `.pr-builder/prompts/*.yaml` format (name, description, template)

---

## 11. Save Prompt Preset Dialog

### Spec Requirements

**Layout:**
- Title: "SAVE PROMPT AS PRESET"
- Preset name input field
- Description input field
- Radio buttons: Project / Global location
- File path preview for each location
- Actions: `Tab Next Field ↑↓ Select Location Enter Confirm Esc Cancel`

**Data:**
- Prompt text (to be saved)
- Preset name
- Description
- Location options (project/global)
- Selected location index

### Current Implementation

**What Exists:**
- ✅ Controller: `SavePromptPreset(name, description, template, location)`
- ✅ Location enum: `PresetLocationProject` / `PresetLocationGlobal`
- ✅ File paths: `.pr-builder/prompts/` (project) or `~/.pr-builder/prompts/` (global)
- ✅ YAML format: name, description, template fields

**What's Missing:**
- ❌ **Dialog**: No TUI dialog for saving prompt presets
- ❌ **Form inputs**: No name/description input fields
- ❌ **Location selection**: No UI to choose project vs global
- ❌ **Path preview**: No display of where preset will be saved

**CLI Equivalents:**
- No CLI command exists (would need to be added)

**Code References:**
- Controller: `internal/controller/controller.go::SavePromptPreset()`
- Domain: `internal/domain/domain.go::PresetLocation`
- YAML format: Controller marshals `{name, description, template}` to YAML

---

## 12. External Editor Screen

### Spec Requirements

**Layout:**
- Title: "EDIT PROMPT TEMPLATE" (same as Edit Prompt Template)
- Message: "Opening in $EDITOR (vim)..."
- Temporary file path displayed
- Instructions: "Save and close the editor to continue."
- Status: "Waiting for editor to close..."

**Data:**
- Editor command (from $EDITOR env var)
- Temporary file path
- Original text (to restore on cancel)

### Current Implementation

**What Exists:**
- ✅ Domain/controller: Prompt text can be set/loaded
- ✅ Session: Prompt stored in session YAML

**What's Missing:**
- ❌ **Screen**: No External Editor screen
- ❌ **Editor integration**: No $EDITOR spawning
- ❌ **Temp file handling**: No temporary file creation/cleanup
- ❌ **Process waiting**: No blocking wait for editor process
- ❌ **Content loading**: No loading edited content back into TUI

**CLI Equivalents:**
- User can manually edit `.pr-builder/prompts/*.yaml` files
- No CLI command to edit prompt interactively

**Code References:**
- None (feature not implemented)

---

## Summary

### Implemented Features
- Domain/controller support for prompts and presets
- Built-in presets (default, detailed, concise, conventional)
- Project/global preset directories
- Preset loading via CLI (`--preset` flag)
- Prompt setting via CLI (`--prompt` flag)

### Missing Features
- **All prompt management UI**: No TUI screens for prompt editing/selection/saving
- **Prompt display**: No UI showing current prompt text
- **Preset browser**: No UI to browse/select presets
- **External editor**: No $EDITOR integration
- **Preset CRUD**: No UI to create/edit/delete presets

### Architectural Notes
- Domain/controller fully support prompt and preset operations
- Preset system mirrors filter preset system (project/global locations, YAML files)
- All prompt management currently done via CLI flags or manual YAML editing
- No TUI integration for prompts (intentional Phase 2 limitation?)
- Prompt is stored in session but not displayed in main screen
