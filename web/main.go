package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/pborman/uuid"
)

type PushEvent interface {
	ProjectKey() string
	Branches() []string
}

type BuildPod struct {
	BuildID             string
	BuildScriptsGitRepo string
	BuildImage          string
	ProjectKey          string
	BranchToBuild       string
	BuildLockKey        string
}

type Handler interface {
	handle(w http.ResponseWriter, r *http.Request)
}

type K8sBase struct {
	MasterURL string
	Locker    Locker
	ApiToken  string
	UserName  string
	Password  string
}

var (
	apiServerBaseURL  = flag.String("api-server-base-url", "https://kubernetes", "Kubernetes API server base URL")
	apiServerUser     = flag.String("api-server-username", "admin", "Kubernetes API server username to use if no service acccount API token is present.")
	apiServerPassword = flag.String("api-server-password", "admin123", "Kubernetes API server password to use if no service acccount API token is present.")
	buildScriptsRepo  = flag.String("build-scripts-repo", "https://github.com/ae6rt/aftomato-build-scripts.git", "Git repo where userland build scripts are held.")
	image             = flag.String("image", "ae6rt/aftomato-build-base:latest", "Build container image.")
	versionFlag       = flag.Bool("version", false, "Print version info and exit.")

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

	httpClient = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
}

func NewK8s(apiServerURL, apiToken, username, password string, locker Locker) K8sBase {
	return K8sBase{
		MasterURL: apiServerURL,
		ApiToken:  apiToken,
		UserName:  username,
		Password:  password,
		Locker:    locker,
	}
}

func (k8s K8sBase) launchBuild(pushEvent PushEvent) error {
	projectKey := pushEvent.ProjectKey()

	buildPod := BuildPod{
		BuildImage:          *image,
		BuildScriptsGitRepo: *buildScriptsRepo,
		ProjectKey:          projectKey,
	}

	tmpl, err := template.New("pod").Parse(podTemplate)
	if err != nil {
		Log.Println(err)
		return err
	}

	for _, branch := range pushEvent.Branches() {
		key := k8s.Locker.Key(projectKey, branch)

		buildPod.BranchToBuild = branch
		buildPod.BuildID = uuid.NewRandom().String()
		buildPod.BuildLockKey = key

		hydratedTemplate := bytes.NewBufferString("")
		err = tmpl.Execute(hydratedTemplate, buildPod)
		if err != nil {
			Log.Println(err)
			continue
		}

		resp, err := k8s.Locker.Lock(key, buildPod.BuildID)
		if err != nil {
			Log.Printf("Failed to acquire lock %s on build %s: %v\n", key, buildPod.BuildID, err)
			continue
		}

		if resp.Node.Value == buildPod.BuildID {
			Log.Printf("Acquired lock on build %s with key %s\n", buildPod.BuildID, key)
			if podError := k8s.createPod(hydratedTemplate.Bytes()); podError != nil {
				Log.Println(podError)
				if _, err := k8s.Locker.Unlock(key, buildPod.BuildID); err != nil {
					Log.Println(err)
				} else {
					Log.Printf("Released lock on build %s with key %s because of pod creation error %v\n", buildPod.BuildID, key, podError)
				}
			}
			Log.Printf("Created pod=%s\n", buildPod.BuildID)
		} else {
			Log.Printf("Failed to acquire lock %s on build %s\n", key, buildPod.BuildID)
		}
	}
	return nil
}

func (base K8sBase) createPod(pod []byte) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/namespaces/default/pods", base.MasterURL), bytes.NewReader(pod))
	if err != nil {
		Log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if base.ApiToken != "" {
		req.Header.Set("Authorization", "Bearer "+base.ApiToken)
	} else {
		req.SetBasicAuth(base.UserName, base.Password)
	}

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
	locker := NewDefaultLock([]string{"http://localhost:2379"})

	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		Log.Printf("No service account token: %v.  Falling back to api server username/password for master authentication.\n", err)
	}

	apiToken := string(data)
	k8s := NewK8s(*apiServerBaseURL, apiToken, *apiServerUser, *apiServerPassword, locker)

	stashHandler := StashHandler{K8sBase: k8s}
	githubHandler := GitHubHandler{K8sBase: k8s}

	http.HandleFunc("/hooks/stash", stashHandler.handle)
	http.HandleFunc("/hooks/github", githubHandler.handle)

	Log.Println("Listening for post-receive messages on port 9090...")
	http.ListenAndServe(":9090", nil)
}
