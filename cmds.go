package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tasks",
	Short: "A CLI task management tool for ~slaying~ your to do list.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return listCmd.RunE(cmd, args)
		}
		return cmd.Help()
	},
}

var addCmd = &cobra.Command{
	Use:   "add NAME",
	Short: "Add a new task with an optional project name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := openDB(setupPath())
		if err != nil {
			return err
		}
		defer t.db.Close()
		project, err := cmd.Flags().GetString("project")
		if err != nil {
			return err
		}
		if err := t.insert(args[0], project); err != nil {
			return err
		}
		fmt.Println("Task successfully added!")
		return nil
	},
}

var whereCmd = &cobra.Command{
	Use:   "where",
	Short: "Show where your tasks are stored",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Println(setupPath())
		return err
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete ID",
	Short: "Delete a task by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := openDB(setupPath())
		if err != nil {
			return err
		}
		defer t.db.Close()
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		err = t.delete(uint(id))
		if err != nil {
			return err
		}
		fmt.Println("Task successfully deleted!")
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update ID",
	Short: "Update a task by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := openDB(setupPath())
		if err != nil {
			return err
		}
		defer t.db.Close()
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		project, err := cmd.Flags().GetString("project")
		if err != nil {
			return err
		}
		prog, err := cmd.Flags().GetInt("status")
		if err != nil {
			return err
		}
		if prog > int(done) {
			return fmt.Errorf("unable to set status: %d", prog)
		}
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		var status string
		switch prog {
		case int(inProgress):
			status = inProgress.String()
		case int(done):
			status = done.String()
		default:
			status = todo.String()
		}
		newTask := task{uint(id), name, project, status, time.Time{}}
		err = t.update(newTask)
		if err != nil {
			return err
		}
		fmt.Println("Task successfully updated!")
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your tasks",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := openDB(setupPath())
		if err != nil {
			return err
		}
		defer t.db.Close()
		tasks, err := t.getTasks()
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			fmt.Println("Add a task to get started")
			return nil
		}
		fmt.Print(setupTable(tasks))
		return nil
	},
}

func setupTable(tasks []task) *table.Table {
	columns := []string{"ID", "NAME", "PROJECT", "STATUS", "CREATED AT"}
	var rows [][]string
	for _, task := range tasks {
		project := task.Project
		if project == "" {
			project = "—"
		}

		rows = append(rows, []string{
			fmt.Sprintf("%d", task.ID),
			task.Name,
			project,
			task.Status,
			task.Created.Format(time.RFC822),
		})
	}
	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Headers(columns...).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Align(lipgloss.Center)
			}
			if row%2 == 0 {
				return lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Padding(0, 3)
			}
			return lipgloss.NewStyle().Padding(0, 3)
		})
	return t
}

func init() {
	addCmd.Flags().StringP(
		"project",
		"p",
		"",
		"specify a project for your task",
	)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	updateCmd.Flags().StringP(
		"name",
		"n",
		"",
		"specify a name for your task",
	)
	updateCmd.Flags().StringP(
		"project",
		"p",
		"",
		"specify a project for your task",
	)
	updateCmd.Flags().IntP(
		"status",
		"s",
		int(todo),
		"specify a status for your task",
	)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(whereCmd)
	rootCmd.AddCommand(deleteCmd)
}
