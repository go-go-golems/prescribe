package layers

import (
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// ExistingCobraFlagsLayer is a wrapper around a standard Glazed schema section
// that prevents it from (re-)adding flags to a cobra.Command.
//
// This is useful when the flags already exist (eg. inherited persistent flags on
// the root command) and adding them again would cause "flag redefined" errors.
//
// The layer still parses values from Cobra using the underlying parameter definitions.
type ExistingCobraFlagsLayer struct {
	*schema.SectionImpl
}

var _ glazed_layers.CobraParameterLayer = &ExistingCobraFlagsLayer{}

func (l *ExistingCobraFlagsLayer) AddLayerToCobraCommand(_ *cobra.Command) error {
	// No-op: flags are expected to already exist on the command (often inherited).
	return nil
}

func (l *ExistingCobraFlagsLayer) Clone() glazed_layers.ParameterLayer {
	cloned := l.SectionImpl.Clone()
	impl, ok := cloned.(*schema.SectionImpl)
	if !ok {
		// This should never happen: schema.NewSection returns *schema.SectionImpl and Clone keeps type.
		return cloned
	}
	return &ExistingCobraFlagsLayer{SectionImpl: impl}
}

func WrapAsExistingCobraFlagsLayer(section schema.Section) (*ExistingCobraFlagsLayer, error) {
	if section == nil {
		return nil, errors.New("section is nil")
	}
	impl, ok := section.(*schema.SectionImpl)
	if !ok {
		return nil, errors.Errorf("section is not a *schema.SectionImpl (got %T)", section)
	}
	return &ExistingCobraFlagsLayer{SectionImpl: impl}, nil
}
