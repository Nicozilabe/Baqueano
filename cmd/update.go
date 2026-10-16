package cmd

import (
	"context"
	"time"

	"github.com/Nicozilabe/baqueano/internal/updater"

	"github.com/spf13/cobra"
)

var (
	updateCheck      bool
	updateForce      bool
	updatePrerelease bool
	updateVersion    string
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update baq to the latest release",
	Long: `Download the latest Baqueano release and replace the current binary.

The right archive for your OS and architecture is fetched from GitHub, checked
against the release checksums and installed in place of the running binary.
Restart baq afterwards to use the new version.`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
		defer cancel()

		_, err := updater.Run(ctx, updater.Options{
			CurrentVersion:    AppVersion,
			TargetVersion:     updateVersion,
			IncludePrerelease: updatePrerelease,
			Force:             updateForce,
			CheckOnly:         updateCheck,
			Out:               cmd.OutOrStdout(),
		})
		return err
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Only check whether an update is available")
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "Reinstall even if already up to date")
	updateCmd.Flags().BoolVar(&updatePrerelease, "pre", false, "Include pre-releases when resolving the latest version")
	updateCmd.Flags().StringVar(&updateVersion, "version", "", "Update to a specific tag (e.g. v1.2.3)")
	RootCmd.AddCommand(updateCmd)
}
