package helper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

// GitHubConfig holds configuration for GitHub operations
type GitHubConfig struct {
	AccessToken string
	AuthorName  string
	AuthorEmail string
	Org         string
	Repo        string
	FilePath    string
	FileContent []byte
	Replace     bool
}

// CalculateHash generates a SHA-256 hash for the given data.
func CalculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// UploadFile uploads or updates a file in a GitHub repository.
func UploadFile(ctx context.Context, config GitHubConfig) (*github.RepositoryContentResponse, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.AccessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Upload File")),
		Content: config.FileContent,
		Branch:  github.String("main"),
		Author: &github.CommitAuthor{
			Name:  github.String(config.AuthorName),
			Email: github.String(config.AuthorEmail),
		},
	}

	// Check if the file exists and handle replacement logic
	existingFile, _, _, err := client.Repositories.GetContents(ctx, config.Org, config.Repo, config.FilePath, nil)
	if err == nil && existingFile != nil && config.Replace {
		// Update existing file
		opts.SHA = github.String(existingFile.GetSHA())
		resp, _, updateErr := client.Repositories.UpdateFile(ctx, config.Org, config.Repo, config.FilePath, opts)
		if updateErr != nil {
			return nil, fmt.Errorf("failed to update file: %w", updateErr)
		}
		return resp, nil
	}

	// Create new file
	resp, _, createErr := client.Repositories.CreateFile(ctx, config.Org, config.Repo, config.FilePath, opts)
	if createErr != nil {
		return nil, fmt.Errorf("failed to create file: %w", createErr)
	}
	return resp, nil
}


// GetFile retrieves the content of a file from a GitHub repository.
func GetFile(ctx context.Context, config GitHubConfig) ([]byte, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.AccessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	fileStream, _, err := client.Repositories.DownloadContents(ctx, config.Org, config.Repo, config.FilePath, nil)
	if err != nil {
		return nil, fmt.Errorf("error downloading file: %w", err)
	}
	defer fileStream.Close()

	content, err := io.ReadAll(fileStream)
	if err != nil {
		return nil, fmt.Errorf("error reading file content: %w", err)
	}

	return content, nil
}
