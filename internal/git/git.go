package git

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jlgore/nsfwctl/pkg/utils"
)

func FetchSlides(repoPath, branchName string) (string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("error opening repository: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("error getting worktree: %v", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return "", fmt.Errorf("error checking out branch: %v", err)
	}

	slidesPath := filepath.Join(repoPath, "slides", "slides.md")
	content, err := ioutil.ReadFile(slidesPath)
	if err != nil {
		return "", fmt.Errorf("error reading slides: %v", err)
	}

	return string(content), nil
}

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

	if err := fetchRepo(repo); err != nil {
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

// func updateRepo(repo *git.Repository) error {
// 	fmt.Println("Updating repository...")
// 	w, err := repo.Worktree()
// 	if err != nil {
// 		return fmt.Errorf("error getting worktree: %v", err)
// 	}

// 	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
// 	if err != nil && err != git.NoErrAlreadyUpToDate {
// 		return fmt.Errorf("error pulling repository: %v", err)
// 	}
// 	return nil
// }

type BranchInfo struct {
	Name        string
	Description string
}

func fetchRepo(repo *git.Repository) error {
	fmt.Println("Fetching updates from remote...")
	err := repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Progress:   nil,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("error fetching repository: %v", err)
	}
	return nil
}

func FetchBranchesWithDescriptions(repoPath string) ([]BranchInfo, error) {
	log.Printf("Starting FetchBranchesWithDescriptions for repo: %s", repoPath)
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Printf("Error opening repository: %v", err)
		return nil, fmt.Errorf("error opening repository: %v", err)
	}

	// Fetch updates before listing branches
	if err := fetchRepo(repo); err != nil {
		log.Printf("Error fetching repository: %v", err)
		return nil, fmt.Errorf("error fetching repository: %v", err)
	}

	log.Print("Fetching remote references")
	remotes, err := repo.Remotes()
	if err != nil {
		log.Printf("Error getting remotes: %v", err)
		return nil, fmt.Errorf("error getting remotes: %v", err)
	}

	var branchInfos []BranchInfo
	for _, remote := range remotes {
		refs, err := remote.List(&git.ListOptions{})
		if err != nil {
			log.Printf("Error listing remote references: %v", err)
			return nil, fmt.Errorf("error listing remote references: %v", err)
		}

		for _, ref := range refs {
			if ref.Name().IsBranch() {
				branchName := ref.Name().Short()
				log.Printf("Processing branch: %s", branchName)
				if utils.IsValidBranchName(branchName) {
					description, _ := getBranchDescription(repo, branchName)
					branchInfo := BranchInfo{
						Name:        branchName,
						Description: description,
					}
					log.Printf("Branch info: %+v", branchInfo)
					branchInfos = append(branchInfos, branchInfo)
				} else {
					log.Printf("Invalid branch name: %s", branchName)
				}
			}
		}
	}

	log.Printf("Total valid branches found: %d", len(branchInfos))
	log.Printf("Returning %d branch infos", len(branchInfos))
	return branchInfos, nil
}

func getBranchDescription(repo *git.Repository, branchName string) (string, error) {
	w, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	// Try to checkout the branch
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewRemoteReferenceName("origin", branchName),
		Force:  true,
	})
	if err != nil {
		// If checkout fails, it might be a remote-only branch
		return "Remote branch - description not available", nil
	}

	descriptionPath := filepath.Join(w.Filesystem.Root(), "description.md")
	file, err := os.Open(descriptionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "No description available", nil
		}
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	lineCount := 0
	maxLines := 5 // Adjust this number to read more or fewer lines

	for scanner.Scan() && lineCount < maxLines {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
			lineCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if len(lines) == 0 {
		return "No description available", nil
	}

	description := strings.Join(lines, "\n")
	if lineCount == maxLines {
		description += "\n..."
	}

	return description, nil
}
