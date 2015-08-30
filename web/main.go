package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/pborman/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
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

type K8sBase struct {
	MasterURL string
	Locker    Locker
	ApiToken  string
}

var (
	apiServerBaseURL           = flag.String("api-server-base-url", "http://localhost:8080", "Kubernetes API server base URL")
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

	httpClient = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
}

func lockKey(projectKey, branch string) string {
	return url.QueryEscape(fmt.Sprintf("%s/%s", projectKey, branch))
}

func NewK8s(apiServerURL string, locker Locker) K8sBase {
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token"); err != nil {
		Log.Printf("Cannot read service account token: %v\n", err)
	} else {
		apiToken = string(data)
	}
	return K8sBase{MasterURL: apiServerURL, Locker: locker}
}

func (k8s K8sBase) build(pushEvent PushEvent) {
	projectKey := pushEvent.ProjectKey()

	buildPod := BuildPod{
		BuildImage:              *image,
		BuildScriptsGitRepo:     *buildScriptsRepo,
		ProjectKey:              projectKey,
		BuildArtifactBucketName: *buildArtifactBucketName,
		ConsoleLogsBucketName:   *buildConsoleLogsBucketName,
	}

	for _, branch := range pushEvent.Branches() {
		buildPod.BranchToBuild = branch
		buildID := uuid.NewRandom().String()
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

		lockKey := lockKey(projectKey, branch)

		resp, err := k8s.Locker.Lock(lockKey, buildID)
		if err != nil {
			Log.Println(err)
			continue
		}

		if resp.Node.Value == buildID {
			Log.Printf("Acquired lock on build %s with key %s\n", buildID, lockKey)
			if podError := k8s.createPod(hydratedTemplate.Bytes()); podError != nil {
				Log.Println(podError)
				if _, err := k8s.Locker.Unlock(lockKey, buildID); err != nil {
					Log.Println(err)
				} else {
					Log.Printf("Released lock on build %s with key %s because of pod creation error %v\n", buildID, lockKey, podError)
				}
			}
			Log.Printf("Created pod=%s\n", buildID)
		}
	}
}

func (base K8sBase) createPod(pod []byte) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/namespaces/default/pods", base.MasterURL), bytes.NewReader(pod))
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
	k8s := NewK8s(*apiServerBaseURL, locker)
	stashHandler := StashHandler{K8sBase: k8s}

	http.HandleFunc("/hooks/stash", stashHandler.handle)
	Log.Printf("Listening for Stash post-receive messages on port 9090.")
	http.ListenAndServe(":9090", nil)
}
