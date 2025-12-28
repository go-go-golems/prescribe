package cmds

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	layersList := []glazed_layers.ParameterLayer{
		repoLayerExisting,
		generationLayer,
	}
	layersList = append(layersList, geppettoLayers...)

	cmdDesc := cmds.NewCommandDescription(
		"generate",
		cmds.WithShort("Generate PR description"),
		cmds.WithLong("Generate a PR description using AI based on the current session."),
		cmds.WithFlags(extraFlags, exportRenderedFlag, printRenderedTokenCountFlag, streamFlag, separatorFlag),
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

	// Output description
	if genSettings.OutputFile != "" {
		if err := os.WriteFile(genSettings.OutputFile, []byte(description), 0644); err != nil {
			return errors.Wrap(err, "failed to write output file")
		}
		fmt.Fprintf(os.Stderr, "Description written to %s\n", genSettings.OutputFile)
	} else {
		fmt.Println(description)
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
