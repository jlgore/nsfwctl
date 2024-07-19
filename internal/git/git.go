package git

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jlgore/nsfwctl/pkg/utils"
)

// EnsureNsfwctlRepo ensures that the nsfwctl repository exists and is up to date
func EnsureNsfwctlRepo(repoURL, branch string) (string, error) {
	appDir, err := utils.GetAppDir()
	if err != nil {
		return "", fmt.Errorf("error getting app directory: %v", err)
	}

	repoDir := filepath.Join(appDir, "infra")

	if err := utils.EnsureDirectory(appDir); err != nil {
		return "", fmt.Errorf("error creating app directory: %v", err)
	}

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return cloneRepo(repoURL, branch, repoDir)
		}
		return "", fmt.Errorf("error opening repository: %v", err)
	}

	if err := updateRepo(repo); err != nil {
		return "", err
	}

	return repoDir, nil
}

func cloneRepo(repoURL, branch, repoDir string) (string, error) {
	fmt.Println("Cloning repository...")
	_, err := git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:           repoURL,
		Progress:      os.Stdout,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return "", fmt.Errorf("error cloning repository: %v", err)
	}
	return repoDir, nil
}

func updateRepo(repo *git.Repository) error {
	fmt.Println("Updating repository...")
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

// FetchBranches retrieves all branches from the repository
func FetchBranches(repoPath string) ([]string, error) {
	log.Printf("Opening repository at: %s", repoPath)
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Printf("Error opening repository: %v", err)
		return nil, fmt.Errorf("error opening repository: %v", err)
	}

	log.Print("Fetching branches")
	branches, err := repo.Branches()
	if err != nil {
		log.Printf("Error fetching branches: %v", err)
		return nil, fmt.Errorf("error fetching branches: %v", err)
	}

	var branchNames []string
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()
		log.Printf("Found branch: %s", branchName)
		if utils.IsValidBranchName(branchName) {
			branchNames = append(branchNames, branchName)
		} else {
			log.Printf("Invalid branch name: %s", branchName)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error iterating branches: %v", err)
		return nil, fmt.Errorf("error iterating branches: %v", err)
	}

	log.Printf("Total valid branches found: %d", len(branchNames))
	return branchNames, nil
}

// SwitchBranch changes the current branch in the repository
func SwitchBranch(repoPath, branchName string) error {
	if !utils.IsValidBranchName(branchName) {
		return fmt.Errorf("invalid branch name: %s", branchName)
	}

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
