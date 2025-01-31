package github

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/mdobak/go-xerrors"
	"github.com/platipy-io/d2s/types"
)

var ErrClient = xerrors.Message("github API call failed")

var httpClient = github.NewClient(&http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 5,
	},
})

type Client struct {
	c *github.Client
}

func NewClient(token string) *Client {
	return &Client{c: httpClient.WithAuthToken(token)}
}

func UserBypass(ctx context.Context) (*types.User, error) {
	return User(ctx, bypassToken)
}

func User(ctx context.Context, token string) (*types.User, error) {
	user, _, err := NewClient(token).c.Users.Get(ctx, "")
	if err != nil {
		return nil, xerrors.New(ErrClient, err)
	}
	return types.NewUser(user.GetName(), user.GetEmail(), token), nil
}

var ErrStarred = xerrors.Message("Github API starred listing failed")

func Starred(ctx context.Context, user *types.User) ([]*types.Repository, error) {
	opts := github.ActivityListStarredOptions{ListOptions: github.ListOptions{}}
	starred, _, err := NewClient(user.Token).c.Activity.ListStarred(ctx, "", &opts)
	if err != nil {
		return nil, xerrors.WithWrapper(ErrStarred, err)
	}
	repos := make([]*types.Repository, len(starred))
	for i, repo := range starred {
		r := types.Repository{
			ID:          *repo.Repository.ID,
			Owner:       *repo.Repository.Owner.Login,
			Name:        *repo.Repository.Name,
			Description: *repo.Repository.Description,
			Language:    *repo.Repository.Language,
			LastUpdated: repo.Repository.UpdatedAt.Time,
		}
		repos[i] = &r
	}
	return repos, nil
}
