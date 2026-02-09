package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var tapGuardCmd = &cobra.Command{
	Use:   "guard",
	Short: "Block forbidden operations (PreToolUse hook)",
	Long: `Block forbidden operations via Claude Code PreToolUse hooks.

Guard commands exit with code 2 to BLOCK tool execution when a policy
is violated. They're called before the tool runs, preventing the
forbidden operation entirely.

Available guards:
  pr-workflow   - Block GitHub PR creation (use MQ instead)

Example hook configuration:
  {
    "PreToolUse": [{
      "matcher": "Bash(gh pr create*)",
      "hooks": [{"command": "gt tap guard pr-workflow"}]
    }]
  }`,
}

var tapGuardPRWorkflowCmd = &cobra.Command{
	Use:   "pr-workflow",
	Short: "Block GitHub PR creation (use MQ instead)",
	Long: `Block GitHub PR creation in Gas Town.

Gas Town workers use feature branches and the merge queue (MQ).
GitHub PRs are not used — Refinery merges from MQ, not from PRs.

This guard blocks:
  - gh pr create

Branch creation (git checkout -b, git switch -c) is allowed because
workers need feature branches for their standard workflow.

Exit codes:
  0 - Operation allowed (not in Gas Town agent context)
  2 - Operation BLOCKED (in agent context)

The guard only blocks when running as a Gas Town agent (crew, polecat,
witness, etc.). Humans running outside Gas Town can still use PRs.`,
	RunE: runTapGuardPRWorkflow,
}

func init() {
	tapCmd.AddCommand(tapGuardCmd)
	tapGuardCmd.AddCommand(tapGuardPRWorkflowCmd)
}

func runTapGuardPRWorkflow(cmd *cobra.Command, args []string) error {
	// Check if we're in a Gas Town agent context
	if !isGasTownAgentContext() {
		// Not in a Gas Town managed context - allow the operation
		return nil
	}

	// We're in a Gas Town context - block PR creation
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "╔══════════════════════════════════════════════════════════════════╗")
	fmt.Fprintln(os.Stderr, "║  ❌ PR CREATION BLOCKED                                          ║")
	fmt.Fprintln(os.Stderr, "╠══════════════════════════════════════════════════════════════════╣")
	fmt.Fprintln(os.Stderr, "║  Gas Town uses the merge queue (MQ), not GitHub PRs.            ║")
	fmt.Fprintln(os.Stderr, "║                                                                  ║")
	fmt.Fprintln(os.Stderr, "║  Instead of:  gh pr create                                      ║")
	fmt.Fprintln(os.Stderr, "║  Do this:     gt done  (submits to MQ, Refinery merges)         ║")
	fmt.Fprintln(os.Stderr, "║                                                                  ║")
	fmt.Fprintln(os.Stderr, "║  Feature branches are fine — use git checkout -b as needed.     ║")
	fmt.Fprintln(os.Stderr, "║  See: ~/gt/docs/PRIMING.md (GUPP principle)                     ║")
	fmt.Fprintln(os.Stderr, "╚══════════════════════════════════════════════════════════════════╝")
	fmt.Fprintln(os.Stderr, "")
	os.Exit(2) // Exit 2 = BLOCK in Claude Code hooks

	return nil
}

// isGasTownAgentContext returns true if we're running as a Gas Town managed agent.
func isGasTownAgentContext() bool {
	// Check environment variables set by Gas Town session management
	envVars := []string{
		"GT_POLECAT",
		"GT_CREW",
		"GT_WITNESS",
		"GT_REFINERY",
		"GT_MAYOR",
		"GT_DEACON",
	}
	for _, env := range envVars {
		if os.Getenv(env) != "" {
			return true
		}
	}

	// Also check if we're in a crew or polecat worktree by path
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	agentPaths := []string{"/crew/", "/polecats/"}
	for _, path := range agentPaths {
		if strings.Contains(cwd, path) {
			return true
		}
	}

	return false
}
