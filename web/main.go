package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/websocket"
)

var (
	apiServerBaseURL       = flag.String("api-server-base-url", "https://kubernetes.default", "Kubernetes API server base URL")
	apiServerUser          = flag.String("api-server-username", "admin", "Kubernetes API server username to use if no service acccount API token is present.")
	apiServerPassword      = flag.String("api-server-password", "admin123", "Kubernetes API server password to use if no service acccount API token is present.")
	awsKey                 = flag.String("aws-access-key", "", "Default decap AWS access key.  /etc/secrets/aws-key in the cluster overrides this.")
	awsSecret              = flag.String("aws-secret-key", "", "Default decap AWS access secret.  /etc/secrets/aws-secret in the cluster overrides this.")
	awsRegion              = flag.String("aws-region", "us-west-1", "Default decap AWS region.  /etc/secrets/aws-region in the cluster overrides this.")
	githubClientID         = flag.String("github-client-id", "", "Default Github ClientID for quering Github repos.  /etc/secrets/github-client-id in the cluster overrides this.")
	githubClientSecret     = flag.String("github-client-secret", "", "Default Github Client Secret for quering Github repos.  /etc/secrets/github-client-secret in the cluster overrides this.")
	buildScriptsRepo       = flag.String("build-scripts-repo", "https://github.com/ae6rt/decap-build-scripts.git", "Git repo where userland build scripts are held.")
	buildScriptsRepoBranch = flag.String("build-scripts-repo-branch", "master", "Branch or revision to use on git repo where userland build scripts are held.")
	versionFlag            = flag.Bool("version", false, "Print version info and exit.")

	Log *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	buildVersion string
	buildCommit  string
	buildDate    string
	buildGoSDK   string
)

func init() {
	flag.Parse()
	Log.Printf("Version: %s, Commit: %s, Date: %s, Go SDK: %s\n", buildVersion, buildCommit, buildDate, buildGoSDK)
	if *versionFlag {
		os.Exit(0)
	}

	*awsKey = kubeSecret("/etc/secrets/aws-key", *awsKey)
	*awsSecret = kubeSecret("/etc/secrets/aws-secret", *awsSecret)
	*awsRegion = kubeSecret("/etc/secrets/aws-region", *awsRegion)
	*githubClientID = kubeSecret("/etc/secrets/github-client-id", *githubClientID)
	*githubClientSecret = kubeSecret("/etc/secrets/github-client-secret", *githubClientSecret)
}

func main() {
	locker := NewDefaultLock([]string{"http://localhost:2379"})
	k8s := NewDefaultDecap(*apiServerBaseURL, *apiServerUser, *apiServerPassword, *awsKey, *awsSecret, *awsRegion, locker)
	awsStorageService := NewAWSStorageService(*awsKey, *awsSecret, *awsRegion)
	scmManagers := map[string]SCMClient{
		"github": NewGithubClient("https://api.github.com", *githubClientID, *githubClientSecret),
	}

	router := httprouter.New()
	router.ServeFiles("/decap/*filepath", http.Dir("./static"))
	router.GET("/api/v1/version", VersionHandler)
	router.GET("/api/v1/projects", ProjectsHandler)
	router.GET("/api/v1/projects/:team/:library/branches", ProjectBranchesHandler(scmManagers))
	router.GET("/api/v1/builds/:team/:library", BuildsHandler(awsStorageService))
	router.DELETE("/api/v1/builds/:id", StopBuildHandler(k8s))
	router.POST("/api/v1/builds/:team/:library", ExecuteBuildHandler(k8s))
	router.GET("/api/v1/teams", TeamsHandler)
	router.GET("/api/v1/logs/:id", LogHandler(awsStorageService))
	router.GET("/api/v1/artifacts/:id", ArtifactsHandler(awsStorageService))
	router.POST("/hooks/:repomanager", HooksHandler(*buildScriptsRepo, *buildScriptsRepoBranch, k8s))

	var err error
	projects, err = assembleProjects(*buildScriptsRepo, *buildScriptsRepoBranch)
	if err != nil {
		Log.Printf("Cannot clone build scripts repository: %v\n", err)
	}
	for _, v := range projects {
		Log.Printf("Project: %+v\n", v)
	}

	go websock()

	Log.Println("decap ready on port 9090...")
	http.ListenAndServe(":9090", router)
}

func kubeSecret(file string, defaultValue string) string {
	if v, err := ioutil.ReadFile(file); err != nil {
		Log.Printf("Secret %s not found in the filesystem.  Using default.\n", file)
		return defaultValue
	} else {
		Log.Printf("Successfully read secret %s from the filesystem\n", file)
		return string(v)
	}

}

func websock() {
	t, err := url.Parse(*apiServerBaseURL)
	if err != nil {
		log.Fatal(err)
	}

	originURL, err := url.Parse(*apiServerBaseURL + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
	if err != nil {
		log.Fatal(err)
	}
	serviceURL, err := url.Parse("wss://" + t.Host + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
	if err != nil {
		log.Fatal(err)
	}

	data, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")

	var hdrs http.Header
	if len(data) == 0 {
		hdrs = map[string][]string{"Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte(*apiServerUser+":"+*apiServerPassword))}}
	} else {
		hdrs = map[string][]string{"Authorization": []string{"Bearer " + string(data)}}
	}

	cfg := websocket.Config{
		Location:  serviceURL,
		Origin:    originURL,
		TlsConfig: &tls.Config{InsecureSkipVerify: true},
		Header:    hdrs,
		Version:   websocket.ProtocolVersionHybi13,
	}

	conn, err := websocket.DialConfig(&cfg)
	if err != nil {
		log.Fatalf("Error opening connection: %v\n", err)
	}
	log.Print("Watching pods on websocket")

	var msg string
	for {
		err := websocket.Message.Receive(conn, &msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("Couldn't receive msg " + err.Error())
			break
		}
		var pod Pod
		if err := json.Unmarshal([]byte(msg), &pod); err != nil {
			log.Println(err)
			continue
		}
		var deletePod bool
		for _, status := range pod.Object.Status.Statuses {
			if status.Name == "build-server" && status.State.Terminated.ContainerID != "" {
				deletePod = true
				break
			}
		}
		if deletePod {
			log.Printf("Would delete:  %+v\n", pod)
		}

	}

}

type Pod struct {
	Object Object `json:"object"`
}

type Object struct {
	Meta   Metadata `json:"metadata"`
	Status Status   `json:"status"`
}

type Metadata struct {
	Name string `json:"name"`
}

type Status struct {
	Statuses []ContainerStatus `json:"containerStatuses"`
}

type ContainerStatus struct {
	Name  string `json:"name"`
	Ready bool   `json:"ready"`
	State State  `json:"state"`
}

type State struct {
	Terminated Terminated `json:"terminated"`
}

type Terminated struct {
	ContainerID string `json:"containerID"`
	ExitCode    int    `json:"exitCode"`
}
