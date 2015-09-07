package githubsdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ae6rt/retry"
)

var Log *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

var httpClient *http.Client = &http.Client{}

type GithubClient struct {
	baseURL      string
	ClientID     string
	ClientSecret string
}

type GithubBranch struct {
	Ref string `json:"ref"`
	URL string `json:"url"`
}

func NewGithubClient(baseURL, clientID, clientSecret string) GithubClient {
	return GithubClient{baseURL: baseURL, ClientID: clientID, ClientSecret: clientSecret}
}

func (gh GithubClient) GetBranches(owner, repository string) ([]GithubBranch, error) {

	branches := make([]GithubBranch, 0)
	var data []byte
	url := fmt.Sprintf("%s/repos/%s/%s/git/refs?client_id=%s&client_secret=%s&page=1", gh.baseURL, owner, repository, gh.ClientID, gh.ClientSecret)
	var response *http.Response

	morePages := true
	for morePages {
		work := func() error {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Accept", "application/json")

			response, err = httpClient.Do(req)
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
		url = nextLink(response.Header.Get("Link"))
		morePages = url == ""
	}

	return branches, nil
}

func nextLink(header string) string {
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
