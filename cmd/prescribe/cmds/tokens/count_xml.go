package tokens

import (
	"context"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	itokens "github.com/go-go-golems/prescribe/internal/tokens"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type CountXMLSettings struct {
	File           string `glazed.parameter:"file"`
	IncludePerFile bool   `glazed.parameter:"per-file"`
	IncludePerItem bool   `glazed.parameter:"per-item"`
}

type CountXMLCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = &CountXMLCommand{}

func NewCountXMLCommand() (*CountXMLCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	fileFlag := parameters.NewParameterDefinition(
		"file",
		parameters.ParameterTypeString,
		parameters.WithHelp("Path to the exported XML-ish file to analyze"),
	)
	perFileFlag := parameters.NewParameterDefinition(
		"per-file",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Include per <file ...>...</file> rows (best-effort)"),
		parameters.WithDefault(true),
	)
	perItemFlag := parameters.NewParameterDefinition(
		"per-item",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Include per <item ...>...</item> rows (best-effort)"),
		parameters.WithDefault(true),
	)

	cmdDesc := cmds.NewCommandDescription(
		"count-xml",
		cmds.WithShort("Count tokens per section in an exported XML-ish payload"),
		cmds.WithLong("Best-effort parser to count tokens per section/tag in an exported XML-ish file (rendered payload or context export)."),
		cmds.WithFlags(fileFlag, perFileFlag, perItemFlag),
		cmds.WithLayersList(
			repoLayerExisting, // accepted for consistency; not used by this command
		),
	)

	return &CountXMLCommand{CommandDescription: cmdDesc}, nil
}

type tagBlock struct {
	Tag        string
	Start      int
	OpenEnd    int // index after the '>' of the opening tag
	CloseStart int // index of '<' of closing tag
	End        int // index after the closing tag
	OpenTag    string
}

func findTagBlocks(s, tag string) []tagBlock {
	openPrefix := "<" + tag
	closeTag := "</" + tag + ">"
	out := []tagBlock{}

	i := 0
	for {
		start := strings.Index(s[i:], openPrefix)
		if start < 0 {
			return out
		}
		start += i

		openEndRel := strings.Index(s[start:], ">")
		if openEndRel < 0 {
			return out
		}
		openEnd := start + openEndRel + 1

		closeStartRel := strings.Index(s[openEnd:], closeTag)
		if closeStartRel < 0 {
			return out
		}
		closeStart := openEnd + closeStartRel
		end := closeStart + len(closeTag)

		openTag := s[start:openEnd]
		out = append(out, tagBlock{
			Tag:        tag,
			Start:      start,
			OpenEnd:    openEnd,
			CloseStart: closeStart,
			End:        end,
			OpenTag:    openTag,
		})
		i = end
	}
}

func getAttr(openTag, key string) string {
	needle := key + "=\""
	i := strings.Index(openTag, needle)
	if i < 0 {
		return ""
	}
	i += len(needle)
	j := strings.Index(openTag[i:], "\"")
	if j < 0 {
		return ""
	}
	return openTag[i : i+j]
}

func findCDATAContent(s, tag string) (string, bool) {
	// Matches patterns emitted by BuildRenderedLLMPayload:
	// <system><![CDATA[ ... ]]></system>
	open := "<" + tag + "><![CDATA["
	closeTag := "]]></" + tag + ">"

	i := strings.Index(s, open)
	if i < 0 {
		return "", false
	}
	i += len(open)
	j := strings.Index(s[i:], closeTag)
	if j < 0 {
		return "", false
	}
	return s[i : i+j], true
}

func (c *CountXMLCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &CountXMLSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to decode count-xml settings")
	}
	if strings.TrimSpace(settings.File) == "" {
		return errors.New("--file is required")
	}

	b, err := os.ReadFile(settings.File)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}
	text := string(b)
	encoding := itokens.EncodingName()

	// Summary row (entire document)
	_ = gp.AddRow(ctx, types.NewRow(
		types.MRP("kind", "document"),
		types.MRP("encoding", encoding),
		types.MRP("file", settings.File),
		types.MRP("tokens", itokens.Count(text)),
		types.MRP("bytes", len(text)),
	))

	// Top-ish sections we care about for prescribe exports.
	for _, tag := range []string{"branches", "commits", "prompt", "files", "context", "llm_payload"} {
		blocks := findTagBlocks(text, tag)
		for idx, blk := range blocks {
			sectionText := text[blk.Start:blk.End]
			row := types.NewRow(
				types.MRP("kind", "section"),
				types.MRP("encoding", encoding),
				types.MRP("section_tag", tag),
				types.MRP("section_index", idx),
				types.MRP("tokens", itokens.Count(sectionText)),
				types.MRP("bytes", len(sectionText)),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	// Rendered payload CDATA (if present)
	if sys, ok := findCDATAContent(text, "system"); ok {
		_ = gp.AddRow(ctx, types.NewRow(
			types.MRP("kind", "cdata"),
			types.MRP("encoding", encoding),
			types.MRP("tag", "system"),
			types.MRP("tokens", itokens.Count(sys)),
			types.MRP("bytes", len(sys)),
		))
	}
	if user, ok := findCDATAContent(text, "user"); ok {
		_ = gp.AddRow(ctx, types.NewRow(
			types.MRP("kind", "cdata"),
			types.MRP("encoding", encoding),
			types.MRP("tag", "user"),
			types.MRP("tokens", itokens.Count(user)),
			types.MRP("bytes", len(user)),
		))
	}

	// Per-file blocks inside <files>...</files> (best-effort)
	if settings.IncludePerFile {
		filesBlocks := findTagBlocks(text, "files")
		for _, fb := range filesBlocks {
			sub := text[fb.OpenEnd:fb.CloseStart]
			fileBlocks := findTagBlocks(sub, "file")
			for _, fblk := range fileBlocks {
				fileText := sub[fblk.Start:fblk.End]
				path := getAttr(fblk.OpenTag, "name")
				ftype := getAttr(fblk.OpenTag, "type")

				innerKind := ""
				innerTokens := 0
				innerBytes := 0

				if diffs := findTagBlocks(fileText, "diff"); len(diffs) > 0 {
					d := diffs[0]
					inner := fileText[d.OpenEnd:d.CloseStart]
					innerKind = "diff"
					innerTokens = itokens.Count(inner)
					innerBytes = len(inner)
				} else if contents := findTagBlocks(fileText, "content"); len(contents) > 0 {
					c := contents[0]
					inner := fileText[c.OpenEnd:c.CloseStart]
					innerKind = "content"
					innerTokens = itokens.Count(inner)
					innerBytes = len(inner)
				}

				row := types.NewRow(
					types.MRP("kind", "file"),
					types.MRP("encoding", encoding),
					types.MRP("path", path),
					types.MRP("file_type", ftype),
					types.MRP("tokens_block", itokens.Count(fileText)),
					types.MRP("bytes_block", len(fileText)),
					types.MRP("inner_kind", innerKind),
					types.MRP("tokens_inner", innerTokens),
					types.MRP("bytes_inner", innerBytes),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
		}
	}

	// Per-item blocks inside <context>...</context> (best-effort)
	if settings.IncludePerItem {
		ctxBlocks := findTagBlocks(text, "context")
		for _, cb := range ctxBlocks {
			sub := text[cb.OpenEnd:cb.CloseStart]
			itemBlocks := findTagBlocks(sub, "item")
			for _, ib := range itemBlocks {
				itemText := sub[ib.Start:ib.End]
				itype := getAttr(ib.OpenTag, "type")
				path := getAttr(ib.OpenTag, "path")

				innerKind := ""
				innerTokens := 0
				innerBytes := 0

				if texts := findTagBlocks(itemText, "text"); len(texts) > 0 {
					t := texts[0]
					inner := itemText[t.OpenEnd:t.CloseStart]
					innerKind = "text"
					innerTokens = itokens.Count(inner)
					innerBytes = len(inner)
				} else if contents := findTagBlocks(itemText, "content"); len(contents) > 0 {
					c := contents[0]
					inner := itemText[c.OpenEnd:c.CloseStart]
					innerKind = "content"
					innerTokens = itokens.Count(inner)
					innerBytes = len(inner)
				}

				row := types.NewRow(
					types.MRP("kind", "item"),
					types.MRP("encoding", encoding),
					types.MRP("item_type", itype),
					types.MRP("path", path),
					types.MRP("tokens_block", itokens.Count(itemText)),
					types.MRP("bytes_block", len(itemText)),
					types.MRP("inner_kind", innerKind),
					types.MRP("tokens_inner", innerTokens),
					types.MRP("bytes_inner", innerBytes),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func NewCountXMLCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewCountXMLCommand()
	if err != nil {
		return nil, err
	}

	cobraCmd, err := cli.BuildCobraCommand(
		glazedCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return nil, err
	}

	return cobraCmd, nil
}
