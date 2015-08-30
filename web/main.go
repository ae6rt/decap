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

	"text/template"

	etcd "github.com/coreos/etcd/client"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
	"time"
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

var (
	buildScriptsRepo           = flag.String("build-scripts-repo", "", "Git repo where userland build scripts are held.")
	buildArtifactBucketName    = flag.String("build-artifact-bucket-name", "aftomato-build-artifacts", "S3 bucket name where build artifacts are stored.")
	buildConsoleLogsBucketName = flag.String("build-console-logs-bucket-name", "aftomato-console-logs", "S3 bucket name where build console logs are stored.")
	image                      = flag.String("image", "", "Build container image.")
	versionFlag                = flag.Bool("version", false, "Print version info and exit.")

	apiToken string

	httpClient *http.Client

	Log *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	etcdConfig etcd.Config

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

	etcdConfig = etcd.Config{
		Endpoints: []string{"http://lockservice:2379"},
		Transport: etcd.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}

}

func stashCommitHandler(w http.ResponseWriter, r *http.Request) {
	Log.Printf("post-receive hook received")
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Printf("%v\n", err)
		w.WriteHeader(200)
		return
	}
	Log.Printf("%s\n", string(data))

	var stashContainer StashContainer
	err = json.Unmarshal(data, &stashContainer)
	if err != nil {
		w.WriteHeader(200)
		return
	}
	w.WriteHeader(200)
	go build(stashContainer)
}

func build(stash PushEvent) {
	// todo anything that arises here from "type bag" is obviously Stash specific.  Generalize this to accomodate Github hooks.
	projectKey := stash.ProjectKey()

	buildPod := BuildPod{
		BuildImage:              *image,
		BuildScriptsGitRepo:     *buildScriptsRepo,
		ProjectKey:              projectKey,
		BuildArtifactBucketName: *buildArtifactBucketName,
		ConsoleLogsBucketName:   *buildConsoleLogsBucketName,
	}

	for _, branch := range stash.Branches() {
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

		resp, err := lockIt(lockKey, buildID)
		if err != nil {
			Log.Println(err)
			continue
		}

		if resp.Node.Value == buildID {
			Log.Printf("Acquired lock on build %s with key %s\n", buildID, lockKey)
			if podError := createPod(hydratedTemplate.Bytes()); podError != nil {
				Log.Println(podError)
				if _, err := unlockIt(lockKey, buildID); err != nil {
					Log.Println(err)
				} else {
					Log.Printf("Released lock on build %s with key %s because of pod creation error %v\n", buildID, lockKey, podError)
				}
			}
			Log.Printf("Created pod=%s\n", buildID)
		}
	}
}

func lockKey(projectKey, branch string) string {
	return url.QueryEscape(fmt.Sprintf("%s/%s", projectKey, branch))
}

func lockIt(lockKey, buildID string) (*etcd.Response, error) {
	c, err := etcd.New(etcdConfig)
	if err != nil {
		return nil, err
	}
	client := etcd.NewKeysAPI(c)
	return client.Set(context.Background(), lockKey, buildID, &etcd.SetOptions{PrevExist: etcd.PrevNoExist})
}

func unlockIt(lockKey, buildID string) (*etcd.Response, error) {
	c, err := etcd.New(etcdConfig)
	if err != nil {
		return nil, err
	}
	client := etcd.NewKeysAPI(c)
	return client.Delete(context.Background(), lockKey, &etcd.DeleteOptions{PrevValue: buildID})
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
	http.HandleFunc("/hooks/stash", stashCommitHandler)
	Log.Printf("Listening for Stash post-receive messages on port 9090.")
	http.ListenAndServe(":9090", nil)
}
