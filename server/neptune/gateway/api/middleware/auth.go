package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"golang.org/x/oauth2"
)

type RequestContextKey string

const (
	UsernameContextKey RequestContextKey = "username"
)

// AdminAuthMiddleware is a somewhat hacky approach to provide authentication by requiring
// a github token is passed in.  Using this token we fetch the authenticated user and validate
// the login against a blessed list of admins.
// There are a couple reasons for this method:
// 1. we need the github username for auditing purposes
// 2. APIs we currently support are clunky and are not GA.
type AdminAuthMiddleware struct {
	Admin valid.Admin
}

func (m *AdminAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return &adminAuthHandler{
		next:  next,
		admin: m.Admin,
	}
}

type adminAuthHandler struct {
	next  http.Handler
	admin valid.Admin
}

func (m *adminAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	ctx := r.Context()

	// get the authenticated user
	client := github.NewClient(
		oauth2.NewClient(ctx,
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
		))

	// empty string will fetch the current authenticated user
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, errors.Wrap(err, "fetching authenticated GH user info"))
		return
	}

	team, _, err := client.Teams.GetTeamMembershipBySlug(ctx, m.admin.GithubTeam.Org, m.admin.GithubTeam.Name, user.GetLogin())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, errors.Wrap(err, "getting admin team membership"))
		return
	}

	if team.GetState() != "active" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "User is not an active admin team member")
		return
	}

	m.next.ServeHTTP(
		w,
		r.WithContext(
			context.WithValue(ctx, UsernameContextKey, user.GetLogin()),
		),
	)
}
