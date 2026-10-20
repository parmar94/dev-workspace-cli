package cli

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the dws version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			v, c := buildVersion()
			fmt.Fprintf(cmd.OutOrStdout(), "dws %s (%s, %s/%s)\n", v, c, runtime.GOOS, runtime.GOARCH)
		},
	}
}

// buildVersion returns the version and commit. Release builds set them with
// -ldflags; `go install …@vX.Y.Z` builds fall back to Go's build info.
func buildVersion() (string, string) {
	v, c := version, commit
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return v, c
	}
	if v == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		v = info.Main.Version
	}
	if c == "none" {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 7 {
				c = s.Value[:7]
			}
		}
	}
	return v, c
}
