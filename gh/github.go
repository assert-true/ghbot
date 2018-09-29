package gh

import (
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
)

type GitHubClientConfig struct {
	Repo string
	User string
	Token string
}

type GitHubClient struct {
	config* GitHubClientConfig
	Ctx context.Context
	client *github.Client
}

func NewGitHubClient(config *GitHubClientConfig) *GitHubClient {
	ctx := context.Background()
	staticTokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken:config.Token})
	oAuthClient := oauth2.NewClient(ctx, staticTokenSource)

	client := github.NewClient(oAuthClient)

	return &GitHubClient{
		config,
		ctx,
		client,
	}
}

func (c *GitHubClient) SetupHook(name, url string) {

	hooks, _, err := c.client.Repositories.ListHooks(c.Ctx, c.config.User, c.config.Repo, nil)
	if err != nil {
		log.Fatalln(err)
	}

	for _, hook := range hooks {
		if *hook.Name == name {
			log.Printf("Hook %s already exists.\n", name)
			_, err := c.client.Repositories.DeleteHook(c.Ctx, c.config.User, c.config.Repo, *hook.ID)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	newHook := &github.Hook{
		Name: &name,
		Config: map[string]interface{}{
			"url": url,
			"content_type": "json",
		},
	}

	_, _, err = c.client.Repositories.CreateHook(c.Ctx, c.config.User, c.config.Repo, newHook)
	if err != nil {
		log.Fatalln(err)
	}
}
