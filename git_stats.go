package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
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
	//token := os.Getenv("GITHUB_AUTH_TOKEN")
	token := "241e304f34f73835dfaf8262a695bfb487c91cfa"
	var client *github.Client

	if token == "" {
		print("!!! No OAuth token. !!!\n\n")
		// End program
	} else {
		tokenService := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
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

func getAuthor(context context.Context, client *github.Client, organization string, repository string) []string {
	commitInfo, resp, err := client.Repositories.ListCommits(context, "Golang-Coach", "Lessons", nil)

	if err != nil {
		fmt.Printf("Problem in commit information %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Page: %d", resp.NextPage)
	var s []string
	var aux string
	// Get the list of all authors of the commits
	for _, v := range commitInfo {
		aux = v.Author.GetLogin()

		if aux == "" {
			aux = v.Commit.Author.GetName()
		}

		// fmt.Printf("index %d, value %+v\n", i, aux)
		s = append(s, aux)
	}

	fmt.Printf("Number of commits: %d\n", len(commitInfo))

	var uniqueAuthors = unique(s)
	return uniqueAuthors
}

func main() {
	context := context.Background()

	client := initClient(context)

	repo, _, err := client.Repositories.Get(context, "Golang-Coach", "Lessons")

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

	/*
		commitInfo, _, err := client.Repositories.ListCommits(context, "Golang-Coach", "Lessons", nil)

		if err != nil {
			fmt.Printf("Problem in commit information %v\n", err)
			os.Exit(1)
		}

		var s []string
		var aux string
		// Get the list of all authors of the commits
		for i, v := range commitInfo {
			aux = v.Author.GetLogin()

			if aux == "" {
				aux = v.Commit.Author.GetName()
			}

			fmt.Printf("index %d, value %+v\n", i, aux)
			s = append(s, commitInfo[i].Author.GetID())
		}
	*/

	var s []string

	s = getAuthor(context, client, "Golang-Coach", "Lessons")

	for i, v := range s {
		fmt.Printf("index %d, value %+v\n", i, v)
	}

}
