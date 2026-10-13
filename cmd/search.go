package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Nicozilabe/baqueano/internal/config"

	"github.com/spf13/cobra"
)

var (
	// outputFormat defines the output format (table, json, simple)
	outputFormat string
	// tagsOnly limits search to tags only
	tagsOnly bool
	// namesOnly limits search to host names only
	namesOnly bool
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search SSH hosts by name, hostname, or tags",
	Long: `Search through your SSH hosts configuration by name, hostname, or tags.
The search is case-insensitive and will match partial strings.

Examples:
  baq search web          # Search for hosts containing "web"
  baq search --tags dev   # Search only in tags for "dev"
  baq search --names prod # Search only in host names for "prod"
  baq search --format json server # Output results in JSON format`,
	Args: cobra.MaximumNArgs(1),
	Run:  runSearch,
}

func runSearch(cmd *cobra.Command, args []string) {
	// Parse SSH configurations
	var hosts []config.SSHHost
	var err error

	if configFile != "" {
		hosts, err = config.ParseSSHConfigFile(configFile)
	} else {
		hosts, err = config.ParseSSHConfig()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading SSH config file: %v\n", err)
		os.Exit(1)
	}

	if len(hosts) == 0 {
		fmt.Println("No SSH hosts found in your configuration file.")
		os.Exit(1)
	}

	// Filter out hidden hosts
	hosts = config.FilterVisibleHosts(hosts)

	// Get search query
	var query string
	if len(args) > 0 {
		query = args[0]
	}

	// Filter hosts based on search criteria
	filteredHosts := filterHosts(hosts, query, tagsOnly, namesOnly)

	// Display results
	if len(filteredHosts) == 0 {
		if query == "" {
			fmt.Println("No hosts found.")
		} else {
			fmt.Printf("No hosts found matching '%s'.\n", query)
		}
		return
	}

	// Output results in specified format
	switch outputFormat {
	case "json":
		outputJSON(filteredHosts)
	case "simple":
		outputSimple(filteredHosts)
	default:
		outputTable(filteredHosts)
	}
}

// filterHosts filters hosts according to the search query and options
func filterHosts(hosts []config.SSHHost, query string, tagsOnly, namesOnly bool) []config.SSHHost {
	var filtered []config.SSHHost

	if query == "" {
		return hosts
	}

	query = strings.ToLower(query)

	for _, host := range hosts {
		matched := false

		// Search in names if not tags-only
		if !tagsOnly {
			// Check the host name
			if strings.Contains(strings.ToLower(host.Name), query) {
				matched = true
			}

			// Check the hostname if not names-only
			if !namesOnly && !matched && strings.Contains(strings.ToLower(host.Hostname), query) {
				matched = true
			}
		}

		// Search in tags if not names-only
		if !namesOnly && !matched {
			for _, tag := range host.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					matched = true
					break
				}
			}
		}

		if matched {
			filtered = append(filtered, host)
		}
	}

	return filtered
}

// outputTable displays results in a formatted table
func outputTable(hosts []config.SSHHost) {
	if len(hosts) == 0 {
		return
	}

	// Calculate column widths
	nameWidth := 4 // "Name"
	hostWidth := 8 // "Hostname"
	userWidth := 4 // "User"
	tagsWidth := 4 // "Tags"

	for _, host := range hosts {
		if len(host.Name) > nameWidth {
			nameWidth = len(host.Name)
		}
		if len(host.Hostname) > hostWidth {
			hostWidth = len(host.Hostname)
		}
		if len(host.User) > userWidth {
			userWidth = len(host.User)
		}
		tagsStr := strings.Join(host.Tags, ", ")
		if len(tagsStr) > tagsWidth {
			tagsWidth = len(tagsStr)
		}
	}

	// Add padding
	nameWidth += 2
	hostWidth += 2
	userWidth += 2
	tagsWidth += 2

	// Print header
	fmt.Printf("%-*s %-*s %-*s %-*s\n", nameWidth, "Name", hostWidth, "Hostname", userWidth, "User", tagsWidth, "Tags")
	fmt.Printf("%s %s %s %s\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", hostWidth),
		strings.Repeat("-", userWidth),
		strings.Repeat("-", tagsWidth))

	// Print hosts
	for _, host := range hosts {
		user := host.User
		if user == "" {
			user = "-"
		}
		tags := strings.Join(host.Tags, ", ")
		if tags == "" {
			tags = "-"
		}
		fmt.Printf("%-*s %-*s %-*s %-*s\n", nameWidth, host.Name, hostWidth, host.Hostname, userWidth, user, tagsWidth, tags)
	}

	fmt.Printf("\nFound %d host(s)\n", len(hosts))
}

// outputSimple displays results in simple format (one per line)
func outputSimple(hosts []config.SSHHost) {
	for _, host := range hosts {
		fmt.Println(host.Name)
	}
}

// outputJSON displays results in JSON format
func outputJSON(hosts []config.SSHHost) {
	type jsonHost struct {
		Name         string   `json:"name"`
		Hostname     string   `json:"hostname"`
		User         string   `json:"user"`
		Port         string   `json:"port"`
		Identity     string   `json:"identity"`
		ProxyJump    string   `json:"proxy_jump"`
		ProxyCommand string   `json:"proxy_command"`
		Options      string   `json:"options"`
		Tags         []string `json:"tags"`
	}

	out := make([]jsonHost, 0, len(hosts))
	for _, host := range hosts {
		tags := host.Tags
		if tags == nil {
			tags = []string{}
		}
		out = append(out, jsonHost{
			Name:         host.Name,
			Hostname:     host.Hostname,
			User:         host.User,
			Port:         host.Port,
			Identity:     host.Identity,
			ProxyJump:    host.ProxyJump,
			ProxyCommand: host.ProxyCommand,
			Options:      host.Options,
			Tags:         tags,
		})
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

func init() {
	// Add search command to root
	RootCmd.AddCommand(searchCmd)

	// Add flags
	searchCmd.Flags().StringVarP(&outputFormat, "format", "f", "table", "Output format (table, json, simple)")
	searchCmd.Flags().BoolVar(&tagsOnly, "tags", false, "Search only in tags")
	searchCmd.Flags().BoolVar(&namesOnly, "names", false, "Search only in host names")
}
