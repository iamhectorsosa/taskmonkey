package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	gap "github.com/muesli/go-app-paths"
)

func setupPath() string {
	scope := gap.NewScope(gap.User, "tasks")

	dirs, err := scope.DataDirs()
	if err != nil {
		log.Fatal(err)
	}

	var taskDir string
	if len(dirs) > 0 {
		taskDir = dirs[0]
	} else {
		taskDir, _ = os.UserHomeDir()
	}

	if err := initTaskDir(taskDir); err != nil {
		log.Fatal(err)
	}

	return taskDir
}

func initTaskDir(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(path, 0o770)
		}
		return err
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
