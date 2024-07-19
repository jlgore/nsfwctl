package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/terraform-exec/tfexec"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type model struct {
	list           list.Model
	repoPath       string
	status         string
	selectedBranch string
}

func initialModel() model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a branch"
	l.SetShowTitle(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return model{
		list:           l,
		status:         "Initializing...",
		selectedBranch: "main",
	}
}

func (m model) Init() tea.Cmd {
	return initializeCmd(m.repoPath)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.selectedBranch = i.title
				return m, switchBranchCmd(m.repoPath, m.selectedBranch)
			}
		}

	case statusMsg:
		m.status = string(msg)

	case branchesMsg:
		items := make([]list.Item, len(msg))
		for i, branch := range msg {
			items[i] = item{title: branch, description: ""}
		}
		m.list.SetItems(items)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return appStyle.Render(fmt.Sprintf(
		"nsfwctl\n\n"+
			"Repository: %s\n"+
			"Current Branch: %s\n"+
			"Status: %s\n\n%s",
		m.repoPath,
		m.selectedBranch,
		m.status,
		m.list.View(),
	))
}

type statusMsg string
type branchesMsg []string
type initMsg string

func initializeCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		err := updateRepo(repoPath)
		if err != nil {
			return statusMsg(fmt.Sprintf("Error updating repo: %v", err))
		}

		initResult := runTerraformInit(repoPath)
		fmt.Println(initResult)

		branches, err := fetchBranches(repoPath)
		if err != nil {
			return statusMsg(fmt.Sprintf("Error fetching branches: %v", err))
		}

		return branchesMsg(branches)
	}
}

func fetchBranches(repoPath string) ([]string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("error opening repository: %v", err)
	}

	branches, err := repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("error fetching branches: %v", err)
	}

	var branchNames []string
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		branchNames = append(branchNames, ref.Name().Short())
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error iterating branches: %v", err)
	}

	return branchNames, nil
}

func switchBranchCmd(repoPath, branchName string) tea.Cmd {
	return func() tea.Msg {
		err := switchBranch(repoPath, branchName)
		if err != nil {
			return statusMsg(fmt.Sprintf("Error switching branch: %v", err))
		}
		return statusMsg(fmt.Sprintf("Switched to branch: %s", branchName))
	}
}

func switchBranch(repoPath, branchName string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("error opening repository: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %v", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return fmt.Errorf("error checking out branch: %v", err)
	}

	return nil
}

func ensureNsfwctlRepo(repoURL, branch string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}

	nsfwctlDir := filepath.Join(homeDir, ".nsfwctl")
	repoDir := filepath.Join(nsfwctlDir, "infra")

	if err := os.MkdirAll(nsfwctlDir, 0755); err != nil {
		return "", fmt.Errorf("error creating .nsfwctl directory: %v", err)
	}

	_, err = git.PlainOpen(repoDir)
	if err != nil {
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
		fmt.Println("Repository already exists. Updating...")
	}

	return repoDir, nil
}

func updateRepo(repoPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("error opening repository: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %v", err)
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("error pulling repository: %v", err)
	}

	return nil
}

func runTerraformInit(repoPath string) string {
	log.Printf("Starting Terraform init process in: %s", repoPath)

	terraformPath, err := exec.LookPath("terraform")
	if err != nil {
		return fmt.Sprintf("terraform not found in PATH: %v", err)
	}
	log.Printf("Terraform executable found at: %s", terraformPath)

	tf, err := tfexec.NewTerraform(repoPath, terraformPath)
	if err != nil {
		return fmt.Sprintf("error creating Terraform object: %v", err)
	}

	logFile := filepath.Join(repoPath, "terraform-init.log")
	f, err := os.Create(logFile)
	if err != nil {
		return fmt.Sprintf("error creating log file: %v", err)
	}
	defer f.Close()

	tf.SetLogger(log.New(f, "", log.Ldate|log.Ltime))

	var stdout, stderr strings.Builder
	tf.SetStdout(&stdout)
	tf.SetStderr(&stderr)

	log.Println("Running Terraform init...")
	err = tf.Init(context.Background(), tfexec.Upgrade(true), tfexec.Reconfigure(true))
	if err != nil {
		log.Printf("Error running terraform init: %v", err)
		return fmt.Sprintf("error running terraform init: %v\nStderr: %s", err, stderr.String())
	}

	log.Println("Terraform init completed successfully")

	terraformDir := filepath.Join(repoPath, ".terraform")
	var terraformContents strings.Builder
	err = filepath.Walk(terraformDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(terraformDir, path)
		if err != nil {
			return err
		}
		terraformContents.WriteString(fmt.Sprintf(".terraform contents: %s\n", rel))
		return nil
	})
	if err != nil {
		log.Printf("Error walking .terraform directory: %v", err)
	}

	providersFile := filepath.Join(repoPath, "providers.tf")
	providersContent, err := ioutil.ReadFile(providersFile)
	if err != nil {
		log.Printf("Error reading providers.tf: %v", err)
	}

	return fmt.Sprintf("Terraform init completed successfully.\n\nInit Output:\n%s\n\nStderr:\n%s\n\n.terraform contents:\n%s\n\nproviders.tf contents:\n%s",
		stdout.String(), stderr.String(), terraformContents.String(), string(providersContent))
}

func main() {
	repoURL := "https://github.com/jlgore/nsfw-infra"
	branch := "main"

	repoPath, err := ensureNsfwctlRepo(repoURL, branch)
	if err != nil {
		log.Fatalf("Failed to ensure repository: %v", err)
	}

	fmt.Printf("Terraform repository is located at: %s\n", repoPath)

	initialState := initialModel()
	initialState.repoPath = repoPath

	p := tea.NewProgram(initialState, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
