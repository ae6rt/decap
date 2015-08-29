package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"text/template"
)

type BuildPod struct {
	BuildID                 string
	BuildScriptsGitRepo     string
	BuildImage              string
	ProjectKey              string
	BranchToBuild           string
	BuildArtifactBucketName string
	ConsoleLogsBucketName   string
}

type bag struct {
	Repository repository  `json:"repository"`
	RefChanges []refChange `json:"refChanges"`
}

type repository struct {
	Slug    string  `json:"slug"`
	Project project `json:"project"`
}

type project struct {
	Key string `json:"key"`
}

type refChange struct {
	RefID string `json:"refId"`
}

var (
	buildScriptsRepo = flag.String("build-scripts-repo", "", "Git repo where userland build scripts are held.")
	image            = flag.String("image", "", "Build container image.")
	versionFlag      = flag.Bool("version", false, "Print version info and exit.")

	Random *os.File

	apiToken string

	httpClient *http.Client

	Log *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	buildInfo string
)

func init() {
	flag.Parse()
	if *versionFlag {
		Log.Printf("%s\n", buildInfo)
		os.Exit(0)
	}
	f, err := os.Open("/dev/urandom")
	if err != nil {
		Log.Fatal(err)
	}
	Random = f

	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		Log.Printf("Cannot read service account token: %v\n", err)
	}
	apiToken = string(data)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient = &http.Client{Transport: tr}
}

func uuid() string {
	b := make([]byte, 16)
	Random.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func handler(w http.ResponseWriter, r *http.Request) {
	Log.Printf("post-receive hook received")
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Printf("%v\n", err)
		w.WriteHeader(200)
		return
	}
	Log.Printf("%s\n", string(data))

	var t bag
	err = json.Unmarshal(data, &t)
	if err != nil {
		w.WriteHeader(200)
		return
	}
	w.WriteHeader(200)
	go build(t)
}

func build(theBag bag) {
	// todo anything that arises here from "type bag" is obviously Stash specific.  Generalize this to accomodate Github hooks.
	projectKey := strings.ToLower(fmt.Sprintf("%s/%s", theBag.Repository.Project.Key, theBag.Repository.Slug))

	buildPod := BuildPod{
		BuildImage:              *image,
		BuildScriptsGitRepo:     *buildScriptsRepo,
		ProjectKey:              projectKey,
		BuildArtifactBucketName: "aftomato-artifacts",
		ConsoleLogsBucketName:   "aftomato-console-logs",
	}

	for _, refID := range theBag.RefChanges {
		branch := strings.ToLower(strings.Replace(refID.RefID, "refs/heads/", "", -1))
		buildPod.BranchToBuild = branch
		buildID := uuid()
		buildPod.BuildID = buildID

		tmpl, err := template.New("pod").Parse(podTemplate)
		if err != nil {
			Log.Println(err)
			continue
		}
		hydratedTemplate := bytes.NewBufferString("")
		err = tmpl.Execute(hydratedTemplate, buildPod)
		if err != nil {
			Log.Println(err)
			continue
		}

		lockKey := fmt.Sprintf("%s/%s", projectKey, branch)

		locked, err := lockBuild(buildID, lockKey)
		if err != nil {
			Log.Printf("Error acquiring lock on build: %v\n", err)
			return
		}
		if !locked {
			Log.Printf("Cannot acquire lock on build\n")
			return
		}
		Log.Println("Acquired lock on build\n")
		createPod(hydratedTemplate.Bytes())
		Log.Printf("Created pod=%s\n", buildID)
	}
}

func createPod(pod []byte) {
	req, err := http.NewRequest("POST", "https://kubernetes/api/v1/namespaces/default/pods", bytes.NewReader(pod))
	if err != nil {
		Log.Println(err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}
	defer func() {
		resp.Body.Close()
	}()

	if resp.StatusCode != 201 {
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			Log.Printf("Error reading non-201 response body: %v\n", err)
		} else {
			Log.Printf("%s\n", string(data))
		}
		return
	}
}

func lockBuild(buildID, key string) (bool, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	data := url.Values{}
	data.Set("value", buildID)

	escapedKey := url.QueryEscape(key)
	Log.Printf("Attempting to acquire a lock on build: %s\n", escapedKey)
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://lockservice:2379/v2/keys/%s?prevExist=false", escapedKey), strings.NewReader(data.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer func() {
		resp.Body.Close()
	}()

	if resp.StatusCode != 201 {
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			Log.Printf("Error reading non-201 response body: %v\n", err)
		} else {
			Log.Printf("%s\n", string(data))
		}
		return false, nil
	}
	return true, nil
}

func unlockBuild(buildID, key string) error {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	escapedKey := url.QueryEscape(key)
	Log.Println("BuildID %s wants to unlock build on key %s\n", buildID, escapedKey)
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://lockservice:2379/v2/keys/%s?prevValue=%s", escapedKey, buildID), nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		resp.Body.Close()
	}()

	if resp.StatusCode != 201 {
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			Log.Printf("Error reading non-201 response body: %v\n", err)
		} else {
			Log.Printf("%s\n", string(data))
		}
		return nil
	}

	return nil
}

func main() {
	http.HandleFunc("/stashhooks", handler)
	Log.Printf("Listening for post-receive messages on port 9090.")
	http.ListenAndServe(":9090", nil)
}
