package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jlgore/nsfwctl/internal/config"
	"github.com/jlgore/nsfwctl/internal/git"
	"github.com/jlgore/nsfwctl/internal/ui"
	"github.com/jlgore/nsfwctl/pkg/utils"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// Setup logging
	logFile, err := setupLogging()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up logging: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Ensure the repository exists and is up to date
	repoPath, err := git.EnsureNsfwctlRepo(config.CurrentConfig.RepoURL, config.CurrentConfig.DefaultBranch)
	if err != nil {
		log.Printf("Failed to ensure repository: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to ensure repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Terraform repository is located at: %s\n", repoPath)

	// Initialize the UI model
	initialState := ui.NewModel(repoPath)

	go git.BackgroundFetch(repoPath) // Start background fetch

	// Run the application
	p := tea.NewProgram(initialState, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func setupLogging() (*os.File, error) {
	appDir, err := utils.GetAppDir()
	if err != nil {
		return nil, fmt.Errorf("error getting app directory: %v", err)
	}

	if err := utils.EnsureDirectory(appDir); err != nil {
		return nil, fmt.Errorf("error creating app directory: %v", err)
	}

	logPath := config.CurrentConfig.LogFile
	if !filepath.IsAbs(logPath) {
		logPath = filepath.Join(appDir, logPath)
	}

	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("error creating log file: %v", err)
	}

	log.SetOutput(logFile)
	return logFile, nil
}
