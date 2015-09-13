package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ae6rt/retry"
)

type RepositoryCoordinates struct {
	Username string
	Password string
	BaseURL  string

	httpClient *http.Client
}

type GithubClient struct {
	RepositoryCoordinates
}

type Branch struct {
	Ref string `json:"ref"`
}

type GithubBranch struct {
	Ref string `json:"ref"`
}

type StashBranches struct {
	IsLastPage    bool          `json:"isLastPage"`
	Size          int           `json:"size"`
	Start         int           `json:"start"`
	NextPageStart int           `json:"nextPageStart"`
	Branch        []StashBranch `json:"values"`
}

type StashBranch struct {
	ID              string `json:"id"`
	DisplayID       string `json:"displayId"`
	LatestChangeSet string `json:"latestChangeset"`
	IsDefault       bool   `json:"isDefault"`
}

type RepositoryClient interface {
	GetBranches(team, repository string) ([]Branch, error)
}

func NewGithubClient(baseURL, clientID, clientSecret string) RepositoryClient {
	return GithubClient{RepositoryCoordinates{BaseURL: baseURL, Username: clientID, Password: clientSecret, httpClient: &http.Client{}}}
}

func (gh GithubClient) GetBranches(owner, repository string) ([]Branch, error) {

	branches := make([]GithubBranch, 0)
	var data []byte
	url := fmt.Sprintf("%s/repos/%s/%s/git/refs?client_id=%s&client_secret=%s&page=1", gh.BaseURL, owner, repository, gh.Username, gh.Password)
	var response *http.Response

	morePages := true
	for morePages {
		work := func() error {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Accept", "application/json")

			response, err = gh.httpClient.Do(req)
			if err != nil {
				return err
			}
			defer func() {
				response.Body.Close()
			}()

			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("git/refs non 200 status code (%d): %s", response.StatusCode, string(data))
			}

			data, err = ioutil.ReadAll(response.Body)
			if err != nil {
				return err
			}

			return nil
		}

		if err := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc).Try(work); err != nil {
			return nil, err
		}

		var b []GithubBranch
		if err := json.Unmarshal(data, &b); err != nil {
			return nil, err
		}

		branches = append(branches, b...)
		url = gh.nextLink(response.Header.Get("Link"))
		morePages = url != ""
	}

	genericBranches := make([]Branch, len(branches))
	for i, v := range branches {
		b := Branch{v.Ref}
		genericBranches[i] = b
	}

	return genericBranches, nil
}

func (gh GithubClient) nextLink(header string) string {
	if header == "" {
		return ""
	}
	links := strings.Split(header, ",")
	for _, v := range links {
		r, _ := regexp.Compile(`^\s*<(.*)>;\s*rel="next"`)
		if r.MatchString(v) {
			s := r.FindStringSubmatch(v)
			return s[1]
		}
	}
	return ""
}
