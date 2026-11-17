package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/skillsproject"
	"github.com/replworks/repl-cli/internal/workspace"
)

func (r *runner) newSkillsCommand() *cobra.Command {
	skills := &cobra.Command{
		Use:   "skills",
		Short: "Install and update Agent Skills",
		Long:  "skills installs and updates the five Agent Skills published at repl.net.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
			}
			return cmd.Help()
		},
	}
	skills.AddCommand(r.newSkillsSyncCommand())

	return skills
}

func (r *runner) newSkillsSyncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Install or update the five Agent Skills",
		Long:  "sync downloads the five Agent Skills from repl.net and installs them byte for byte under .claude/skills/ of the project directory, updating existing files.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := r.projectDir()
			if err != nil {
				return err
			}

			outcome := skillsproject.Coordinator{
				Source:    r.skillSource(),
				Workspace: workspace.GrantWrite(workspace.New(dir)),
			}.Run()
			r.status = renderSkillsSync(cmd, outcome)

			return nil
		},
	}
}

func renderSkillsSync(cmd *cobra.Command, outcome skillsproject.Outcome) int {
	out := cmd.OutOrStdout()

	for _, item := range outcome.Items {
		_, _ = fmt.Fprintf(out, "%s: %s: %s\n", item.Name, item.Path, item.State)
	}

	if len(outcome.Errors) > 0 {
		for _, err := range outcome.Errors {
			_, _ = fmt.Fprintln(out, errorLine(err))
		}
		return 2
	}

	return 0
}
