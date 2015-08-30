package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type PushEvent interface {
	ProjectKey() string
	Branches() []string
}

type BuildPod struct {
	BuildID                 string
	BuildScriptsGitRepo     string
	BuildImage              string
	ProjectKey              string
	BranchToBuild           string
	BuildArtifactBucketName string
	ConsoleLogsBucketName   string
}

type Handler interface {
	handle(w http.ResponseWriter, r *http.Request)
}

var (
	buildScriptsRepo           = flag.String("build-scripts-repo", "", "Git repo where userland build scripts are held.")
	buildArtifactBucketName    = flag.String("build-artifact-bucket-name", "aftomato-build-artifacts", "S3 bucket name where build artifacts are stored.")
	buildConsoleLogsBucketName = flag.String("build-console-logs-bucket-name", "aftomato-console-logs", "S3 bucket name where build console logs are stored.")
	image                      = flag.String("image", "", "Build container image.")
	versionFlag                = flag.Bool("version", false, "Print version info and exit.")

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

	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token"); err != nil {
		Log.Printf("Cannot read service account token: %v\n", err)
	} else {
		apiToken = string(data)
	}

	httpClient = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
}

func lockKey(projectKey, branch string) string {
	return url.QueryEscape(fmt.Sprintf("%s/%s", projectKey, branch))
}

func createPod(pod []byte) error {
	req, err := http.NewRequest("POST", "https://kubernetes/api/v1/namespaces/default/pods", bytes.NewReader(pod))
	if err != nil {
		Log.Println(err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		resp.Body.Close()
	}()

	if resp.StatusCode != 201 {
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			Log.Printf("Error reading non-201 response body: %v\n", err)
			return err
		} else {
			Log.Printf("%s\n", string(data))
			return nil
		}
	}
	return nil
}

func main() {
	locker := NewDefaultLock([]string{"http://lockservice:2379"})
	stashHandler := StashHandler{Locker: locker}

	http.HandleFunc("/hooks/stash", stashHandler.handle)
	Log.Printf("Listening for Stash post-receive messages on port 9090.")
	http.ListenAndServe(":9090", nil)
}
