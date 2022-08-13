package main

import (
	"context"
	"flag"
	"os"

	"github.com/gjkim42/actions-runner-cleaner/cleaner"
	"github.com/golang/glog"
	"github.com/google/go-github/v45/github"
)

var (
	baseURL    string
	uploadURL  string
	username   = os.Getenv("GITHUB_USERNAME")
	secret     = os.Getenv("GITHUB_SECRET")
	org        string
	repository string
)

func init() {
	flag.Set("v", "0")
	flag.Set("logtostderr", "true")
	flag.StringVar(&baseURL, "base-url", baseURL, "Base URL for GitHub API")
	flag.StringVar(&uploadURL, "upload-url", uploadURL, "Upload URL for GitHub API")
	flag.StringVar(&org, "org", org, "GitHub organization to clean up runners")
	flag.StringVar(&repository, "repository", repository, "GitHub repository to clean up runners")
}

func main() {
	flag.Parse()

	if username == "" || secret == "" {
		glog.Error("GITHUB_USERNAME and GITHUB_PASSWORD must be set")
		os.Exit(1)
	}

	client, err := newGithubClient()
	if err != nil {
		glog.Error(err)
		os.Exit(1)
	}

	err = cleaner.NewCleanerWithRepository(client, org, repository).Run(context.Background())
	if err != nil {
		glog.Error(err)
		os.Exit(1)
	}
}

func newGithubClient() (*github.Client, error) {
	trans := &github.BasicAuthTransport{
		Username: username,
		Password: secret,
	}

	if baseURL != "" && uploadURL != "" {
		return github.NewEnterpriseClient(baseURL, uploadURL, trans.Client())
	}

	return github.NewClient(trans.Client()), nil
}
