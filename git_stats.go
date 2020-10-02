package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// Model
type GE struct {
	Enabler string
	Owner   string
	Repo    string
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

func getForkedRepos(context context.Context, client *github.Client, organization string, repository string) ([]string, []string) {
	var owner []string // List of owners of the different repositories, including main owner and all forked ones
	var repo []string  // List of repo name of the different repositories, including main owner and all forked ones

	/*opt := &RepositoryListForksOptions{
		ListOptions: ListOptions{PerPage: 30},
	}*/

	owner = append(owner, organization)
	repo = append(repo, repository)

	// Need to check if there is more than 30 forks in a repo
repeat:
	forks, _, err := client.Repositories.ListForks(context, organization, repository, nil)

	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			fmt.Println("GitHub rate limit reached, waiting 1h")
			time.Sleep(61 * time.Minute)
			goto repeat
		} else {
			fmt.Printf("Problems getting branches information %v\n", err)
			os.Exit(1)
		}
	}

	for _, v := range forks {
		owner = append(owner, v.Owner.GetLogin())
		repo = append(repo, v.GetName())
	}

	return owner, repo
}

func getAuthors(context context.Context, client *github.Client, organization string, repository string) []string {
	// For each branch, we need to get the unique list of commits authors
	var s []string
	var aux string

	// Initialize the CommitsLisOptions parameter
	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	owner, repo := getForkedRepos(context, client, organization, repository)

	for i, u := range owner {
		t := repo[i]

		fmt.Printf("\nOwner: %+v      Repo: %+v\n", u, t)

		// First, for the main repo extract all branches and get the owners of all commits
		// Need to check if there is more than 30 branches in one repo...
	repeat1:
		branches, _, err := client.Repositories.ListBranches(context, u, t, nil)

		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				fmt.Println("GitHub rate limit reached, waiting 1h")
				time.Sleep(61 * time.Minute)
				goto repeat1
			} else {
				fmt.Printf("Problems getting branches information %v\n", err)
				os.Exit(1)
			}
		}

		for _, v := range branches {
			fmt.Printf("    Branch sha: %+v        name: %+v\n", v.Commit.GetSHA(), *v.Name)

			opt.SHA = v.Commit.GetSHA()
			opt.Page = 0

			for {
			repeat2:
				commitInfo, resp_commits, err := client.Repositories.ListCommits(context, u, t, opt)

				if err != nil {
					if _, ok := err.(*github.RateLimitError); ok {
						fmt.Println("GitHub rate limit reached, waiting 1h")
						time.Sleep(61 * time.Minute)
						goto repeat2
					} else {
						fmt.Printf("Problems getting branches information %v\n", err)
						os.Exit(1)
					}
				}

				// Get the list of all commit authors
				for _, w := range commitInfo {
					aux = w.Author.GetLogin()

					if aux == "" {
						aux = w.Commit.Author.GetName()
					}

					// fmt.Printf("index %d, value %+v\n", i, aux)
					s = append(s, aux)
				}

				if resp_commits.NextPage == 0 {
					break
				}

				opt.Page = resp_commits.NextPage

				s = unique(s)
			}

		}

		s = unique(s)
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

func getRepoData(context context.Context, client *github.Client, owner string, repo string) string {
repeat:
	repoData, _, err := client.Repositories.Get(context, owner, repo)

	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			fmt.Println("GitHub rate limit reached, waiting 1h")
			time.Sleep(61 * time.Minute)
			goto repeat
		} else {
			fmt.Printf("Problems getting branches information %v\n", err)
			os.Exit(1)
		}
	}

	return *repoData.FullName
}

func getRepos() []GE {
	var ge []GE

	jsonFile, err := os.Open("enablers.json")

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened enablers.json")

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &ge)

	return ge
}

func main() {
	var authors []string

	// we iterate through every user within our users array and
	// print out the user Type, their name, and their facebook url
	// as just an example
	ge := getRepos()

	for i := 0; i < len(ge); i++ {
		var owner = ge[i].Owner
		var repo = ge[i].Repo

		fmt.Println("Enabler name: " + ge[i].Enabler + "    Owner: " + owner + "    Repo: " + repo)

		context := context.Background()

		client := initClient(context)

		repoName := getRepoData(context, client, owner, repo)
		fmt.Printf("\n%+v\n\n", repoName)

		authors = append(authors, getAuthors(context, client, owner, repo)...)
		authors = unique(authors)
	}

	fmt.Printf("\nResults obtained (different users)\n")
	for i, v := range authors {
		fmt.Printf("    Index %d, value %+v\n", i, v)
	}

}
