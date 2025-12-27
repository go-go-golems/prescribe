# Tasks

## TODO

- [ ] Add tasks here

- [x] Phase 1: Create RepositoryLayer in prescribe/pkg/layers/repository.go with schema.NewSection() and fields.New() API
- [x] Phase 1: Create SessionLayer in prescribe/pkg/layers/session.go
- [x] Phase 1: Create FilterLayer in prescribe/pkg/layers/filter.go
- [x] Phase 1: Create GenerationLayer in prescribe/pkg/layers/generation.go
- [x] Phase 1: Create helper functions GetRepositorySettings(), GetSessionSettings(), GetFilterSettings(), GetGenerationSettings() for extracting settings from parsed layers
- [x] Phase 1: Update root command to integrate Glazed help system (repo/target remain root persistent flags; commands parse them via existing-flags wrapper)
- [x] Phase 1: Create controller initialization helpers that use layers instead of reading Cobra flags directly
- [x] Phase 2: Port filter list command to Glazed output (no dual-mode / no back-compat)
- [x] Phase 2: Port filter show command to Glazed output (no dual-mode / no back-compat)
- [x] Phase 2: Port filter test command to Glazed output (no dual-mode / no back-compat)
- [x] Phase 2: Port session show command to Glazed output (no dual-mode / no back-compat)
- [x] Phase 3: Port filter add command to Glazed BareCommand
- [ ] Phase 3: Port generate command to Glazed (decide BareCommand vs GlazeCommand output contract)
- [x] Phase 4: Port filter remove command to Glazed BareCommand
- [x] Phase 4: Port filter clear command to Glazed BareCommand
- [x] Phase 4: Port session init command to Glazed BareCommand
- [x] Phase 4: Port session load command to Glazed BareCommand
- [x] Phase 4: Port session save command to Glazed BareCommand
- [x] Phase 4: Port file toggle command to Glazed BareCommand
- [x] Phase 4: Port context add command to Glazed BareCommand
- [ ] Phase 4: Update tui command to use layers (no structural changes needed)
- [ ] Testing: Create unit tests for all layer creation and settings extraction
- [ ] Testing: Create integration tests for commands with mock controllers
- [ ] Testing: Create E2E tests for full command execution with real git repositories
- [ ] Testing: Verify behavior for ported commands (no backwards compatibility promised; update scripts as needed)
- [ ] Documentation: Update command help text and examples to reflect Glazed integration
- [ ] Documentation: Add examples showing Glazed structured output usage (JSON/YAML/CSV) for ported query commands
- [x] Phase 1: Update root command to add PersistentPreRunE with logging.InitLoggerFromCobra(cmd) for logging initialization
- [x] Phase 1: Add logging layer to root command using logging.AddLoggingLayerToRootCommand(rootCmd, "prescribe") in main()
- [x] Phase 1: Set up Glazed help system with help.NewHelpSystem() and help_cmd.SetupCobraRootCommand() in main()
- [x] Phase 1: Update main() function structure to follow Glazed program initialization pattern (logging, help system, command registration)
