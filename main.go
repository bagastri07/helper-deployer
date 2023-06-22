package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var validEnvironments = []string{"staging", "deva", "devb"}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "helper",
		Short: "A CLI helper tool",
	}

	var deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Clone current branch and push to remote git",
		Run:   deploy,
	}

	var environment string

	deployCmd.Flags().StringVarP(&environment, "name", "n", "", "Deployment environment")

	rootCmd.AddCommand(deployCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func deploy(cmd *cobra.Command, args []string) {
	environment, err := cmd.Flags().GetString("name")
	if err != nil {
		log.Fatal(err)
	}

	if !isValidEnvironment(environment) {
		fmt.Println("Invalid environment")
		os.Exit(1)
	}

	// Get Git username
	gitName := getGitUserName()

	// Get repository name
	repoName := getRepositoryName()

	// Generate branch name with timestamp
	branch := generateBranchName(environment)

	// Clone the current branch
	fmt.Printf("Deployer: %s\n", gitName)
	fmt.Printf("Repository: %s\n", repoName)
	fmt.Println("Cloning branch:", branch)
	err = runCommand("git", "clone", ".", branch)
	if err != nil {
		log.Fatal(err)
	}

	// Push the cloned branch to the remote git
	fmt.Println("Pushing branch:", branch)
	err = runCommand("git", "push", "origin", branch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Deployment completed successfully!")
}

func isValidEnvironment(environment string) bool {
	for _, env := range validEnvironments {
		if env == environment {
			return true
		}
	}
	return false
}

func generateBranchName(environment string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	return environment + "-" + timestamp
}

func getGitUserName() string {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(output))
}

func getRepositoryName() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	repoPath := filepath.Join(dir, ".git")
	repo, err := exec.Command("git", "--git-dir="+repoPath, "remote", "show", "origin").Output()
	if err != nil {
		log.Fatal(err)
	}
	return extractRepoName(string(repo))
}

func extractRepoName(remoteOutput string) string {
	lines := strings.Split(remoteOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Fetch URL:") {
			urlParts := strings.Split(line, "/")
			return strings.TrimSpace(urlParts[len(urlParts)-1])
		}
	}
	return ""
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
