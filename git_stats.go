package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// Model
type Package struct {
	FullName      string
	Description   string
	StarsCount    int
	ForksCount    int
	LastUpdatedBy string
}

func initClient(context context.Context) *github.Client {
	githubAPIKey, exists := os.LookupEnv("GITHUB_API_KEY")

	var client *github.Client

	if exists {
		tokenService := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubAPIKey})
		tokenClient := oauth2.NewClient(context, tokenService)

		client = github.NewClient(tokenClient)
	}

	return client
}

func unique(authors []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range authors {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func getAuthors(context context.Context, client *github.Client, organization string, repository string) []string {
	// For each branch, we need to get the unique list of commits authors
	// List branches	GET	/repos/{owner}/{repo}/branches{?page}	Done	Done
	// Get branch		GET	/repos/{owner}/{repo}/branches/{branch}	Done	Done
	var s []string
	var aux string

	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 30},
	}

	branches, _, err := client.Repositories.ListBranches(context, organization, repository, nil)

	if err != nil {
		fmt.Printf("Problems getting branches information %v\n", err)
		os.Exit(1)
	}

	for _, v := range branches {
		fmt.Printf("Branch sha: %+v\n", v.Commit.GetSHA())

		opt.SHA = v.Commit.GetSHA()
		opt.Page = 0

		for {
			commitInfo, resp_commits, err := client.Repositories.ListCommits(context, organization, repository, opt)

			if err != nil {
				fmt.Printf("Problem in commit information %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Page: %d", resp_commits.NextPage)

			// Get the list of all commit authors
			for _, v := range commitInfo {
				aux = v.Author.GetLogin()

				if aux == "" {
					aux = v.Commit.Author.GetName()
				}

				// fmt.Printf("index %d, value %+v\n", i, aux)
				s = append(s, aux)
			}

			fmt.Printf("  Number of commits: %d\n", len(commitInfo))

			if resp_commits.NextPage == 0 {
				break
			}

			opt.Page = resp_commits.NextPage

			s = unique(s)
		}

	}

	var uniqueAuthors = unique(s)
	return uniqueAuthors
}

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	context := context.Background()

	client := initClient(context)

	repo, _, err := client.Repositories.Get(context, "flopezag", "github-api-testing")

	if err != nil {
		fmt.Printf("Problem in getting repository information %v\n", err)
		os.Exit(1)
	}

	pack := &Package{
		FullName:    *repo.FullName,
		Description: *repo.Description,
		ForksCount:  *repo.ForksCount,
		StarsCount:  *repo.StargazersCount,
	}

	fmt.Printf("\n%+v\n\n", pack.FullName)

	var s []string

	s = getAuthors(context, client, "flopezag", "github-api-testing")

	for i, v := range s {
		fmt.Printf("index %d, value %+v\n", i, v)
	}

}
