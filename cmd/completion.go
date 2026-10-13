package cmd

import (
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for baq.

To load completions:

Bash:
  $ source <(baq completion bash)
  
  # To load completions for each session, add to your ~/.bashrc:
  # echo 'source <(baq completion bash)' >> ~/.bashrc

Zsh:
  $ source <(baq completion zsh)
  
  # To load completions for each session, add to your ~/.zshrc:
  # echo 'source <(baq completion zsh)' >> ~/.zshrc

Fish:
  $ baq completion fish | source
  
  # To load completions for each session:
  $ baq completion fish > ~/.config/fish/completions/baq.fish

PowerShell:
  PS> baq completion powershell | Out-String | Invoke-Expression
  
  # To load completions for each session, add to your PowerShell profile:
  # Add-Content $PROFILE 'baq completion powershell | Out-String | Invoke-Expression'
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletionV2(out, true)
		case "zsh":
			return cmd.Root().GenZshCompletion(out)
		case "fish":
			return cmd.Root().GenFishCompletion(out, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(out)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
