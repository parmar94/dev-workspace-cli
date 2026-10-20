package cli

import (
	"github.com/parmar94/dev-workspace-cli/internal/config"
	"github.com/parmar94/dev-workspace-cli/internal/registry"
)

// isInteractive reports whether prompts can be shown.
func isInteractive() bool { return false }

// promptSelection asks for any values not set by flags.
func promptSelection(*selection, *config.Config, *registry.Registry, func(string) bool) error {
	return nil
}
