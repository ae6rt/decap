package githubsdk

import "testing"

func TestLinks(t *testing.T) {

	header := `<https://api.github.com/repositories/20580498/git/refs?client_id=cid&client_secret=sekrit&page=2>; rel="next", <https://api.github.com/repositories/20580498/git/refs?client_id=cid&client_secret=sekrit&page=393>; rel="last`
	s := nextLink(header)
	if s != "https://api.github.com/repositories/20580498/git/refs?client_id=cid&client_secret=sekrit&page=2" {
		t.Fatalf("Want https://api.github.com/repositories/20580498/git/refs?client_id=cid&client_secret=sekrit&page=2  but got %s\n", s)
	}

	w := nextLink("")
	if w != "" {
		t.Fatalf("Want empty string but got %s\n", w)
	}
}

func TestActualProject(t *testing.T) {
	/*
		client := NewGithubClient("https://api.github.com", "cid", "sekrit")
		b, err := client.GetBranches("gorilla", "mux")
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range b {
			fmt.Printf("%+v\n", v)
		}
	*/
}
