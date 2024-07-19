package terraform

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
)

// InitTerraform initializes Terraform in the given directory
func InitTerraform(repoPath string) (string, error) {
	log.Printf("Starting Terraform init process in: %s", repoPath)

	terraformPath, err := exec.LookPath("terraform")
	if err != nil {
		return "", fmt.Errorf("terraform not found in PATH: %v", err)
	}
	log.Printf("Terraform executable found at: %s", terraformPath)

	tf, err := tfexec.NewTerraform(repoPath, terraformPath)
	if err != nil {
		return "", fmt.Errorf("error creating Terraform object: %v", err)
	}

	logFile := filepath.Join(repoPath, "terraform-init.log")
	f, err := os.Create(logFile)
	if err != nil {
		return "", fmt.Errorf("error creating log file: %v", err)
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
		return "", fmt.Errorf("error running terraform init: %v\nStderr: %s", err, stderr.String())
	}

	log.Println("Terraform init completed successfully")

	return formatTerraformOutput(repoPath, stdout.String(), stderr.String())
}

func formatTerraformOutput(repoPath, stdout, stderr string) (string, error) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Terraform init completed successfully.\n\nInit Output:\n%s\n\nStderr:\n%s\n\n", stdout, stderr))

	// List .terraform directory contents
	terraformDir := filepath.Join(repoPath, ".terraform")
	err := filepath.Walk(terraformDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(terraformDir, path)
		if err != nil {
			return err
		}
		output.WriteString(fmt.Sprintf(".terraform contents: %s\n", rel))
		return nil
	})
	if err != nil {
		log.Printf("Error walking .terraform directory: %v", err)
	}

	// Read providers.tf content
	providersFile := filepath.Join(repoPath, "providers.tf")
	providersContent, err := os.ReadFile(providersFile)
	if err != nil {
		log.Printf("Error reading providers.tf: %v", err)
	} else {
		output.WriteString(fmt.Sprintf("\nproviders.tf contents:\n%s", string(providersContent)))
	}

	return output.String(), nil
}

// PlanTerraform runs terraform plan and returns the plan output
func PlanTerraform(repoPath string) (string, error) {
	tf, err := tfexec.NewTerraform(repoPath, "terraform")
	if err != nil {
		return "", fmt.Errorf("error creating Terraform object: %v", err)
	}

	var stdout, stderr strings.Builder
	tf.SetStdout(&stdout)
	tf.SetStderr(&stderr)

	log.Println("Running Terraform plan...")
	_, err = tf.Plan(context.Background())
	if err != nil {
		return "", fmt.Errorf("error running terraform plan: %v\nStderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// ApplyTerraform runs terraform apply
func ApplyTerraform(repoPath string) (string, error) {
	tf, err := tfexec.NewTerraform(repoPath, "terraform")
	if err != nil {
		return "", fmt.Errorf("error creating Terraform object: %v", err)
	}

	var stdout, stderr strings.Builder
	tf.SetStdout(&stdout)
	tf.SetStderr(&stderr)

	log.Println("Running Terraform apply...")
	err = tf.Apply(context.Background())
	if err != nil {
		return "", fmt.Errorf("error running terraform apply: %v\nStderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
