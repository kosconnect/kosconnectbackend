package config

import "os"

func GetGitHubToken() string {
    return os.Getenv("GH_ACCESS_TOKEN")
}
