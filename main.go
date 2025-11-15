package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Repo struct {
	Name  string `json:"name"`
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`
}

type Commit struct {
	Commit struct {
		Author struct {
			Name string `json:"name"`
		} `json:"author"`
	} `json:"commit"`
}

type Issue struct {
	Title string `json:"title"`
	Repo  struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./github-activity <username>")
		os.Exit(1)
	}
	username := os.Args[1]

	client := &http.Client{Timeout: 20 * time.Second}

	reposURL := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100", username)
	reposResp, err := client.Get(reposURL)
	if err != nil {
		fmt.Println("❌ Request failed:", err)
		return
	}
	defer reposResp.Body.Close()

	if reposResp.StatusCode != 200 {
		fmt.Println("❌ GitHub API returned:", reposResp.Status)
		return
	}

	var repos []Repo
	if err := json.NewDecoder(reposResp.Body).Decode(&repos); err != nil {
		fmt.Println("❌ Failed to parse JSON:", err)
		return
	}

	for _, repo := range repos {
		commitsURL := fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/commits?author=%s&per_page=100",
			repo.Owner.Login, repo.Name, username,
		)
		commitsResp, err := client.Get(commitsURL)
		if err != nil {
			continue
		}
		defer commitsResp.Body.Close()

		if commitsResp.StatusCode != 200 {
			continue
		}

		var commits []Commit
		if err := json.NewDecoder(commitsResp.Body).Decode(&commits); err != nil {
			continue
		}

		if len(commits) > 0 {
			fmt.Printf("Pushed %d commits to %s/%s\n", len(commits), repo.Owner.Login, repo.Name)
		}
	}

	issuesURL := fmt.Sprintf("https://api.github.com/search/issues?q=author:%s+type:issue&per_page=100", username)
	issuesResp, err := client.Get(issuesURL)
	if err == nil {
		defer issuesResp.Body.Close()
		if issuesResp.StatusCode == 200 {
			var result struct {
				Items []struct {
					RepositoryURL string `json:"repository_url"`
				} `json:"items"`
			}
			if err := json.NewDecoder(issuesResp.Body).Decode(&result); err == nil {
				for _, item := range result.Items {
					var ownerRepo string
					fmt.Sscanf(item.RepositoryURL, "https://api.github.com/repos/%s", &ownerRepo)
					fmt.Printf("Opened a new issue in %s\n", ownerRepo)
				}
			}
		}
	}

	starredURL := fmt.Sprintf("https://api.github.com/users/%s/starred?per_page=100", username)
	starredResp, err := client.Get(starredURL)
	if err == nil {
		defer starredResp.Body.Close()
		if starredResp.StatusCode == 200 {
			var starred []Repo
			if err := json.NewDecoder(starredResp.Body).Decode(&starred); err == nil {
				for _, repo := range starred {
					fmt.Printf("Starred %s/%s\n", repo.Owner.Login, repo.Name)
				}
			}
		}
	}
}
