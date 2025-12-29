package cmds

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	geppettolayers "github.com/go-go-golems/geppetto/pkg/layers"
	gepsettings "github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/glazed/pkg/appconfig"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	glazed_config "github.com/go-go-golems/glazed/pkg/config"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	papi "github.com/go-go-golems/prescribe/internal/api"
	pexport "github.com/go-go-golems/prescribe/internal/export"
	"github.com/go-go-golems/prescribe/internal/git"
	"github.com/go-go-golems/prescribe/internal/github"
	"github.com/go-go-golems/prescribe/internal/prdata"
	"github.com/go-go-golems/prescribe/internal/tokens"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var generateCmd *cobra.Command

type GenerateCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &GenerateCommand{}

type GenerateExtraSettings struct {
	ExportContext           bool   `glazed.parameter:"export-context"`
	ExportRendered          bool   `glazed.parameter:"export-rendered"`
	PrintRenderedTokenCount bool   `glazed.parameter:"print-rendered-token-count"`
	Stream                  bool   `glazed.parameter:"stream"`
	Separator               string `glazed.parameter:"separator"`
	Create                  bool   `glazed.parameter:"create"`
	CreateDryRun            bool   `glazed.parameter:"create-dry-run"`
	CreateDraft             bool   `glazed.parameter:"create-draft"`
	CreateBase              string `glazed.parameter:"create-base"`
}

func NewGenerateCommand() (*GenerateCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	generationLayer, err := prescribe_layers.NewGenerationLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create generation layer")
	}

	geppettoLayers, err := geppettolayers.CreateGeppettoLayers()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create geppetto parameter layers")
	}

	extraFlags := parameters.NewParameterDefinition(
		"export-context",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Print the full generation context (prompt + files + context) and exit (no inference)"),
		parameters.WithDefault(false),
	)
	exportRenderedFlag := parameters.NewParameterDefinition(
		"export-rendered",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Print the rendered LLM payload (system+user) and exit (no inference)"),
		parameters.WithDefault(false),
	)
	printRenderedTokenCountFlag := parameters.NewParameterDefinition(
		"print-rendered-token-count",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Print token counts for the rendered LLM payload (system+user) to stderr (no inference required)"),
		parameters.WithDefault(false),
	)
	streamFlag := parameters.NewParameterDefinition(
		"stream",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Stream inference output/events to stderr while still producing a final result"),
		parameters.WithDefault(false),
	)
	separatorFlag := parameters.NewParameterDefinition(
		"separator",
		parameters.ParameterTypeString,
		parameters.WithHelp("Separator format for export flags: xml (default), markdown, simple, begin-end, default"),
		parameters.WithDefault("xml"),
	)
	createFlag := parameters.NewParameterDefinition(
		"create",
		parameters.ParameterTypeBool,
		parameters.WithHelp("After generating, create the PR via GitHub CLI (gh pr create)"),
		parameters.WithDefault(false),
	)
	createDryRunFlag := parameters.NewParameterDefinition(
		"create-dry-run",
		parameters.ParameterTypeBool,
		parameters.WithHelp("With --create: print the create actions (git push + gh ...) but do not execute them"),
		parameters.WithDefault(false),
	)
	createDraftFlag := parameters.NewParameterDefinition(
		"create-draft",
		parameters.ParameterTypeBool,
		parameters.WithHelp("With --create: create the PR as a draft"),
		parameters.WithDefault(false),
	)
	createBaseFlag := parameters.NewParameterDefinition(
		"create-base",
		parameters.ParameterTypeString,
		parameters.WithHelp("With --create: base branch for the PR"),
		parameters.WithDefault("main"),
	)

	layersList := []glazed_layers.ParameterLayer{
		repoLayerExisting,
		generationLayer,
	}
	layersList = append(layersList, geppettoLayers...)

	cmdDesc := cmds.NewCommandDescription(
		"generate",
		cmds.WithShort("Generate PR description"),
		cmds.WithLong("Generate a PR description using AI based on the current session."),
		cmds.WithFlags(extraFlags, exportRenderedFlag, printRenderedTokenCountFlag, streamFlag, separatorFlag, createFlag, createDryRunFlag, createDraftFlag, createBaseFlag),
		cmds.WithLayersList(
			layersList...,
		),
	)

	return &GenerateCommand{CommandDescription: cmdDesc}, nil
}

func (c *GenerateCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	genSettings, err := prescribe_layers.GetGenerationSettings(parsedLayers)
	if err != nil {
		return err
	}

	extra := &GenerateExtraSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, extra); err != nil {
		return errors.Wrap(err, "failed to decode generate extra settings")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	// Prefer loading an existing default session so this command reflects current state.
	helpers.LoadDefaultSessionIfExists(ctrl)

	// Load session if specified (overrides default session).
	if genSettings.LoadSession != "" {
		if err := ctrl.LoadSession(genSettings.LoadSession); err != nil {
			return errors.Wrap(err, "failed to load session")
		}
		fmt.Fprintf(os.Stderr, "Loaded session from: %s\n", genSettings.LoadSession)
	}

	// Override prompt if specified
	if genSettings.Prompt != "" {
		ctrl.SetPrompt(genSettings.Prompt, nil)
	} else if genSettings.Preset != "" {
		if err := ctrl.LoadPromptPreset(genSettings.Preset); err != nil {
			return errors.Wrap(err, "failed to load preset")
		}
	}

	// Override PR title/description if specified (takes precedence over session.yaml).
	if strings.TrimSpace(genSettings.Title) != "" {
		ctrl.GetData().Title = genSettings.Title
	}
	if strings.TrimSpace(genSettings.Description) != "" {
		ctrl.GetData().Description = genSettings.Description
	}

	// Export-only path (no inference).
	if extra.ExportContext && extra.ExportRendered {
		return errors.New("flags --export-context and --export-rendered are mutually exclusive")
	}
	if extra.ExportContext || extra.ExportRendered {
		req, err := ctrl.BuildGenerateDescriptionRequest()
		if err != nil {
			return err
		}
		sep := pexport.SeparatorType(extra.Separator)

		if extra.PrintRenderedTokenCount {
			sys, user, err := papi.CompilePrompt(req)
			if err != nil {
				return err
			}
			sysTokens := tokens.Count(sys)
			userTokens := tokens.Count(user)
			fmt.Fprintf(os.Stderr, "Rendered payload token counts (encoding=%s): system=%d user=%d total=%d\n", tokens.EncodingName(), sysTokens, userTokens, sysTokens+userTokens)

			renderedExport, err := pexport.BuildRenderedLLMPayload(req, sep)
			if err == nil {
				fmt.Fprintf(os.Stderr, "Rendered payload export token count (separator=%s): %d\n", extra.Separator, tokens.Count(renderedExport))
			}
		}

		text := ""
		if extra.ExportRendered {
			rendered, err := pexport.BuildRenderedLLMPayload(req, sep)
			if err != nil {
				return err
			}
			text = rendered
		} else {
			text = pexport.BuildGenerationContext(req, sep)
		}
		if genSettings.OutputFile != "" {
			if err := os.WriteFile(genSettings.OutputFile, []byte(text), 0644); err != nil {
				return errors.Wrap(err, "failed to write output file")
			}
			what := "Context"
			if extra.ExportRendered {
				what = "Rendered payload"
			}
			fmt.Fprintf(os.Stderr, "%s written to %s\n", what, genSettings.OutputFile)
			return nil
		}
		fmt.Print(text)
		return nil
	}

	// Optional debug output (no inference required): rendered payload token counts.
	if extra.PrintRenderedTokenCount {
		req, err := ctrl.BuildGenerateDescriptionRequest()
		if err != nil {
			return err
		}
		sep := pexport.SeparatorType(extra.Separator)
		sys, user, err := papi.CompilePrompt(req)
		if err != nil {
			return err
		}
		sysTokens := tokens.Count(sys)
		userTokens := tokens.Count(user)
		fmt.Fprintf(os.Stderr, "Rendered payload token counts (encoding=%s): system=%d user=%d total=%d\n", tokens.EncodingName(), sysTokens, userTokens, sysTokens+userTokens)

		renderedExport, err := pexport.BuildRenderedLLMPayload(req, sep)
		if err == nil {
			fmt.Fprintf(os.Stderr, "Rendered payload export token count (separator=%s): %d\n", extra.Separator, tokens.Count(renderedExport))
		}
	}

	// StepSettings parsing happens here (higher up), then injected into API service.
	stepSettings, err := gepsettings.NewStepSettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return errors.Wrap(err, "failed to build AI step settings from parsed layers")
	}
	ctrl.SetStepSettings(stepSettings)

	// Generate description
	fmt.Fprintf(os.Stderr, "Generating PR description...\n")
	description := ""
	if extra.Stream {
		description, err = ctrl.GenerateDescriptionStreaming(ctx, os.Stderr)
	} else {
		description, err = ctrl.GenerateDescription(ctx)
	}
	if err != nil {
		return errors.Wrap(err, "failed to generate description")
	}

	if extra.Stream {
		// Print a deterministic end-of-run summary to stderr so users can see the parsed structure
		// even if they streamed raw deltas. Keep stdout reserved for the final description output.
		data := ctrl.GetData()
		if data != nil && data.GeneratedPRData != nil {
			b, err := yaml.Marshal(data.GeneratedPRData)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n--- Parsed PR data (failed to marshal): %v ---\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "\n--- Parsed PR data (YAML) ---\n%s\n", string(b))
			}
		} else if data != nil && data.GeneratedPRDataParseError != "" {
			fmt.Fprintf(os.Stderr, "\n--- Parsed PR data: failed (%s) ---\n", data.GeneratedPRDataParseError)
		} else {
			fmt.Fprintf(os.Stderr, "\n--- Parsed PR data: not available ---\n")
		}
	}

	// Persist last generated structured PR data (for `prescribe create --use-last`).
	data := ctrl.GetData()
	if data != nil && data.GeneratedPRData != nil {
		if repoSettings, err := prescribe_layers.GetRepositorySettings(parsedLayers); err == nil {
			path := prdata.LastGeneratedPRDataPath(repoSettings.RepoPath)
			if err := prdata.WriteGeneratedPRDataToYAMLFile(path, data.GeneratedPRData); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to write last generated PR data: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Parsed PR data written to %s\n", path)
			}
		}
	}

	// Output description
	if genSettings.OutputFile != "" {
		if err := os.WriteFile(genSettings.OutputFile, []byte(description), 0644); err != nil {
			return errors.Wrap(err, "failed to write output file")
		}
		fmt.Fprintf(os.Stderr, "Description written to %s\n", genSettings.OutputFile)
	} else {
		fmt.Println(description)
	}

	// Optional: create PR from parsed structured data.
	if extra.Create {
		repoSettings, err := prescribe_layers.GetRepositorySettings(parsedLayers)
		if err != nil {
			return err
		}
		data := ctrl.GetData()
		if data == nil || data.GeneratedPRData == nil {
			return errors.New("--create requires parsed PR data (GeneratedPRData), but it was not available")
		}

		opts := github.CreatePROptions{
			Title: data.GeneratedPRData.Title,
			Body:  data.GeneratedPRData.Body,
			Base:  extra.CreateBase,
			Draft: extra.CreateDraft,
		}
		args, err := github.BuildGhCreatePRArgs(opts)
		if err != nil {
			return err
		}

		if extra.CreateDryRun {
			fmt.Fprintln(os.Stderr, "generate --create-dry-run: would push branch and create PR via GitHub CLI:")
			fmt.Fprintf(os.Stderr, "  repo: %s\n", repoSettings.RepoPath)
			fmt.Fprintf(os.Stderr, "  command: git push\n")
			fmt.Fprintf(os.Stderr, "  command: gh %s\n", strings.Join(github.RedactGhArgs(args), " "))
			return nil
		}

		fmt.Fprintln(os.Stderr, "generate --create: pushing branch and creating PR...")
		gitSvc, err := git.NewService(repoSettings.RepoPath)
		if err != nil {
			return err
		}
		if err := gitSvc.PushCurrentBranch(ctx); err != nil {
			return err
		}

		ghSvc := github.NewService(repoSettings.RepoPath)
		out, err := ghSvc.CreatePR(ctx, opts)
		if err != nil {
			// Save PR data for manual retry.
			failPath := prdata.FailurePRDataPath(repoSettings.RepoPath, time.Now())
			if werr := prdata.WriteGeneratedPRDataToYAMLFile(failPath, data.GeneratedPRData); werr == nil {
				fmt.Fprintf(os.Stderr, "generate --create: saved PR data to %s\n", failPath)
			}
			return err
		}
		fmt.Fprintln(os.Stderr, out)
	}

	return nil
}

func InitGenerateCmd() error {
	glazedCmd, err := NewGenerateCommand()
	if err != nil {
		return err
	}

	// Build a middleware chain that supports:
	// - config files (prescribe config) for parameter defaults / overrides
	// - PINOCCHIO profiles.yaml (bootstrap parse of profile selection + profile loading)
	// - env overrides (PRESCRIBE_* and PINOCCHIO_*)
	//
	// The precedence is:
	// defaults < profiles < config < env < args < flags
	generateMiddlewares := func(parsedCommandLayers *glazed_layers.ParsedLayers, cmd *cobra.Command, args []string) ([]cmd_middlewares.Middleware, error) {
		// 1) Resolve config files (low -> high precedence).
		commandSettings := &cli.CommandSettings{}
		if parsedCommandLayers != nil {
			_ = parsedCommandLayers.InitializeStruct(cli.CommandSettingsSlug, commandSettings)
		}

		var configFiles []string
		// Base config discovery (if present).
		if p, err := glazed_config.ResolveAppConfigPath("prescribe", ""); err == nil && p != "" {
			configFiles = append(configFiles, p)
		}
		// Optional explicit config overlay.
		if commandSettings.ConfigFile != "" {
			configFiles = append(configFiles, commandSettings.ConfigFile)
		}
		// Optional "load parameters from file" overlay (legacy).
		if commandSettings.LoadParametersFromFile != "" {
			configFiles = append(configFiles, commandSettings.LoadParametersFromFile)
		}

		// Optional: Load Pinocchio config as a *defaults overlay* (lower precedence than profiles).
		//
		// This is useful because many users keep common AI defaults (like ai-max-response-tokens)
		// in `~/.pinocchio/config.yaml`, but `prescribe` is a separate app name and therefore
		// won't discover it via `ResolveAppConfigPath("prescribe", "")`.
		//
		// We intentionally apply this AFTER profiles (lower precedence) so profiles can still
		// select provider/model without being overridden by global defaults.
		pinocchioConfigFile := ""
		if p, err := glazed_config.ResolveAppConfigPath("pinocchio", ""); err == nil && p != "" {
			pinocchioConfigFile = p
		}

		// 2) Bootstrap-parse profile selection using appconfig (circularity-safe).
		// This allows `profile-settings.profile` and `profile-settings.profile-file` to be set by:
		// - cobra flags (--profile/--profile-file)
		// - env vars (PINOCCHIO_PROFILE / PINOCCHIO_PROFILE_FILE)
		// - config files under `profile-settings:`
		type bootstrap struct {
			Profile cli.ProfileSettings
		}
		profileSettingsLayer, err := cli.NewProfileSettingsLayer()
		if err != nil {
			return nil, err
		}
		bootstrapParser, err := appconfig.NewParser[bootstrap](
			appconfig.WithDefaults(),
			appconfig.WithConfigFiles(configFiles...),
			appconfig.WithEnv("PINOCCHIO"),
			appconfig.WithCobra(cmd, args),
		)
		if err != nil {
			return nil, err
		}
		if err := bootstrapParser.Register(appconfig.LayerSlug(cli.ProfileSettingsSlug), profileSettingsLayer, func(t *bootstrap) any {
			return &t.Profile
		}); err != nil {
			return nil, err
		}
		boot, err := bootstrapParser.Parse()
		if err != nil {
			return nil, err
		}

		xdgConfigPath, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		defaultProfileFile := filepath.Join(xdgConfigPath, "pinocchio", "profiles.yaml")
		profileName := boot.Profile.Profile
		if profileName == "" {
			profileName = "default"
		}
		profileFile := boot.Profile.ProfileFile
		if profileFile == "" {
			profileFile = defaultProfileFile
		}

		profileMiddleware := cmd_middlewares.GatherFlagsFromProfiles(
			defaultProfileFile,
			profileFile,
			profileName,
			"default",
			parameters.WithParseStepSource("profiles"),
			parameters.WithParseStepMetadata(map[string]interface{}{
				"profileFile": profileFile,
				"profile":     profileName,
			}),
		)

		// 3) Main chain (highest -> lowest precedence in slice order).
		middlewares_ := []cmd_middlewares.Middleware{
			cmd_middlewares.ParseFromCobraCommand(cmd, parameters.WithParseStepSource("cobra")),
			cmd_middlewares.GatherArguments(args, parameters.WithParseStepSource("arguments")),

			// Environment overrides
			cmd_middlewares.UpdateFromEnv("PRESCRIBE", parameters.WithParseStepSource("env")),
			cmd_middlewares.UpdateFromEnv("PINOCCHIO", parameters.WithParseStepSource("env")),

			// Config files: low -> high precedence
			cmd_middlewares.LoadParametersFromFiles(configFiles),

			// Profiles: apply after defaults but before config/env/flags
			profileMiddleware,
		}
		if pinocchioConfigFile != "" {
			middlewares_ = append(middlewares_,
				cmd_middlewares.LoadParametersFromFile(
					pinocchioConfigFile,
					// Pinocchio config often includes non-layer top-level keys like `repositories: [...]`.
					// The default loader expects every top-level key to be a layer map, so we filter here.
					cmd_middlewares.WithConfigFileMapper(func(raw interface{}) (map[string]map[string]interface{}, error) {
						out := map[string]map[string]interface{}{}
						rm, ok := raw.(map[string]interface{})
						if !ok {
							return out, nil
						}
						for k, v := range rm {
							vm, ok := v.(map[string]interface{})
							if !ok {
								continue
							}
							out[k] = vm
						}
						return out, nil
					}),
					cmd_middlewares.WithParseOptions(
						parameters.WithParseStepSource("pinocchio-config"),
						parameters.WithParseStepMetadata(map[string]interface{}{
							"config_file": pinocchioConfigFile,
						}),
					),
				),
			)
		}
		// Defaults (lowest precedence)
		middlewares_ = append(middlewares_,
			cmd_middlewares.SetFromDefaults(parameters.WithParseStepSource(parameters.SourceDefaults)),
		)
		return middlewares_, nil
	}

	cobraCmd, err := cli.BuildCobraCommand(
		glazedCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			EnableProfileSettingsLayer: true,
			MiddlewaresFunc:            generateMiddlewares,
		}),
	)
	if err != nil {
		return err
	}

	generateCmd = cobraCmd
	return nil
}
