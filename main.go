package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	list     list.Model
	tf       *tfexec.Terraform
	choice   string
	quitting bool
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func initialModel() model {
	items := []list.Item{
		item{title: "Network ACLs", desc: "Deploy Network Access Control Lists"},
		item{title: "Security Groups", desc: "Configure Security Groups"},
		item{title: "VPN Gateway", desc: "Set up a VPN Gateway"},
		item{title: "WAF", desc: "Deploy Web Application Firewall"},
		item{title: "Apply All", desc: "Deploy all security controls"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "AWS VPC Security Controls"

	return model{list: l}
}

func (m model) Init() tea.Cmd {
	return nil
}

func runTerraformInit(repoPath string) error {
	// Ensure Terraform is installed and get its path
	terraformPath, err := ensureTerraform()
	if err != nil {
		return fmt.Errorf("failed to ensure Terraform is installed: %v", err)
	}

	// Create a new Terraform object
	tf, err := tfexec.NewTerraform(repoPath, terraformPath)
	if err != nil {
		return fmt.Errorf("error creating Terraform object: %v", err)
	}

	// Set working directory to where your Terraform configurations are
	workingDir := filepath.Join(repoPath, "path/to/terraform/configs")
	tf.SetWorkingDir(workingDir)

	// Run terraform init
	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return fmt.Errorf("error running terraform init: %v", err)
	}

	fmt.Println("Terraform init completed successfully")
	return nil
}

func ensureNsfwctlRepo(repoURL, branch string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}

	nsfwctlDir := filepath.Join(homeDir, ".nsfwctl")
	repoDir := filepath.Join(nsfwctlDir, "terraform-repo")

	// Ensure .nsfwctl directory exists
	if err := os.MkdirAll(nsfwctlDir, 0755); err != nil {
		return "", fmt.Errorf("error creating .nsfwctl directory: %v", err)
	}

	// Check if the repository already exists
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		// Repository doesn't exist, clone it
		fmt.Println("Cloning repository...")
		_, err = git.PlainClone(repoDir, false, &git.CloneOptions{
			URL:           repoURL,
			Progress:      os.Stdout,
			ReferenceName: plumbing.NewBranchReferenceName(branch),
		})
		if err != nil {
			return "", fmt.Errorf("error cloning repository: %v", err)
		}
	} else {
		// Repository exists, pull latest changes
		fmt.Println("Updating repository...")
		w, err := repo.Worktree()
		if err != nil {
			return "", fmt.Errorf("error getting worktree: %v", err)
		}

		err = w.Pull(&git.PullOptions{
			RemoteName:    "origin",
			ReferenceName: plumbing.NewBranchReferenceName(branch),
			Progress:      os.Stdout,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return "", fmt.Errorf("error pulling repository: %v", err)
		}
	}

	fmt.Println("Repository is up to date.")
	return repoDir, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i.title
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Bye!\n"
	}
	if m.choice != "" {
		return fmt.Sprintf("You chose %s!\n", m.choice)
	}
	return docStyle.Render(m.list.View())
}

func main() {

	repoURL := "https://github.com/jlgore/nsfw-infra.git"
	branch := "main"

	repoPath, err := ensureNsfwctlRepo(repoURL, branch)
	if err != nil {
		log.Fatalf("Failed to ensure repository: %v", err)
	}

	fmt.Printf("Terraform repository is located at: %s\n", repoPath)
	// Initialize Terraform
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.9.1")),
	}
	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}

	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Error getting current user: %v", err)
	}
	// Get the home directory
	homeDir := currentUser.HomeDir

	// Define the path for the .nsfwctl folder
	nsfwctlDir := filepath.Join(homeDir, ".nsfwctl")

	// Check if the directory exists
	_, err = os.Stat(nsfwctlDir)
	if os.IsNotExist(err) {
		// Directory doesn't exist, so create it
		err = os.Mkdir(nsfwctlDir, 0755)
		if err != nil {
			log.Fatalf("Error creating directory: %v", err)
		}
		fmt.Printf("Created directory: %s\n", nsfwctlDir)
	} else if err != nil {
		// Some other error occurred
		log.Fatalf("Error checking directory: %v", err)
	} else {
		fmt.Printf("Directory already exists: %s\n", nsfwctlDir)
	}

	tf, err := tfexec.NewTerraform(nsfwctlDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	// Run Bubble Tea
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	// Process the user's choice
	if m, ok := m.(model); ok {
		switch m.choice {
		case "Network ACLs":
			// Apply Network ACLs Terraform module
		case "Security Groups":
			// Apply Security Groups Terraform module
		case "VPN Gateway":
			// Apply VPN Gateway Terraform module
		case "WAF":
			// Apply WAF Terraform module
		case "Apply All":
			// Apply all modules
		}
	}
}
