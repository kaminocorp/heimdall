package codebase

import "context"

// GitHub implements the codebase connector using the GitHub API.
type GitHub struct {
	// TODO: add repo URL, access token, HTTP client
}

func NewGitHub() *GitHub {
	return &GitHub{}
}

func (g *GitHub) Connect(ctx context.Context) error {
	return nil
}

func (g *GitHub) Health(ctx context.Context) error {
	// TODO: verify repo access
	return nil
}

func (g *GitHub) Query(ctx context.Context, query string) (any, error) {
	// TODO: search codebase via GitHub API
	return nil, nil
}

func (g *GitHub) Close() error {
	return nil
}
