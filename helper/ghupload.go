package ghupload

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	// "errors"
	"fmt"
	// "io"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

// GitHubConfig holds configuration for GitHub operations
type GitHubConfig struct {
	Token  string
	Owner  string
	Repo   string
	Branch string
	Folder string
}

// UploadFile uploads a file to a GitHub repository, updating the file if it already exists.
func UploadFile(ctx context.Context, config GitHubConfig, path string, content []byte) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Calculate SHA-256 hash of the content
	hash := sha256.Sum256(content)
	checksum := hex.EncodeToString(hash[:])

	// Check if the file already exists in the repository
	existingFile, _, _, err := client.Repositories.GetContents(ctx, config.Owner, config.Repo, path, &github.RepositoryContentGetOptions{
		Ref: config.Branch,
	})

	var fileContent github.RepositoryContentFileOptions
	fileContent.Message = github.String(fmt.Sprintf("Upload or update file with hash: %s", checksum))
	fileContent.Content = content
	fileContent.Branch = github.String(config.Branch)

	if err == nil && existingFile != nil {
		// File exists, add SHA to update the file
		sha := existingFile.GetSHA()
		fileContent.SHA = github.String(sha)
	} else if err != nil {
		// Log error for non-existent file (optional)
		fmt.Printf("File not found, creating a new one: %v\n", err)
	}

	// Create or update the file
	_, _, err = client.Repositories.CreateFile(ctx, config.Owner, config.Repo, path, &fileContent)
	if err != nil {
		return fmt.Errorf("failed to upload or update file: %w", err)
	}

	return nil
}

// GetFile retrieves a file from a GitHub repository
func GetFile(ctx context.Context, config GitHubConfig, path string) ([]byte, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	fileContent, _, _, err := client.Repositories.GetContents(ctx, config.Owner, config.Repo, path, &github.RepositoryContentGetOptions{
		Ref: config.Branch,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode file content: %w", err)
	}

	return []byte(content), nil
}

// Example: Function to handle file upload from form-data
func createBoardingHouse(w http.ResponseWriter, r *http.Request) {
	config := GitHubConfig{
		Token:  "your-github-token",
		Owner:  "your-github-username",
		Repo:   "your-repo-name",
		Branch: "main",
	}

	r.ParseMultipartForm(10 << 20) // 10 MB limit

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to process file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file content", http.StatusInternalServerError)
		return
	}

	filePath := fmt.Sprintf("boardinghouse/%s", handler.Filename)
	ctx := context.Background()

	err = UploadFile(ctx, config, filePath, content)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload file: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "File uploaded successfully")
}

// Example: Function to handle file retrieval
func getBoardingHouseFile(w http.ResponseWriter, r *http.Request) {
	config := GitHubConfig{
		Token:  "your-github-token",
		Owner:  "your-github-username",
		Repo:   "your-repo-name",
		Branch: "main",
	}

	filePath := r.URL.Query().Get("filePath")
	if filePath == "" {
		http.Error(w, "filePath query parameter is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	content, err := GetFile(ctx, config, filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func main() {
	http.HandleFunc("/uploadBoardingHouse", createBoardingHouse)
	http.HandleFunc("/getBoardingHouseFile", getBoardingHouseFile)

	http.ListenAndServe(":8080", nil)
}
