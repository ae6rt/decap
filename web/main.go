package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
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
	noWebsocket            = flag.Bool("no-websocket", false, "Do not start websocket client that watches pods.")
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
	k8s := NewDefaultDecap(*apiServerBaseURL, *apiServerUser, *apiServerPassword, *awsKey, *awsSecret, *awsRegion, locker, *buildScriptsRepo, *buildScriptsRepoBranch)
	awsStorageService := NewAWSStorageService(*awsKey, *awsSecret, *awsRegion)
	scmManagers := map[string]SCMClient{
		"github": NewGithubClient("https://api.github.com", *githubClientID, *githubClientSecret),
	}

	router := httprouter.New()
	router.ServeFiles("/decap/*filepath", http.Dir("./static"))
	router.GET("/api/v1/version", VersionHandler)
	router.GET("/api/v1/projects", ProjectsHandler)
	router.GET("/api/v1/projects/:team/:project/refs", ProjectRefsHandler(scmManagers))
	router.GET("/api/v1/builds/:team/:project", BuildsHandler(awsStorageService))
	router.DELETE("/api/v1/builds/:id", StopBuildHandler(k8s))
	router.POST("/api/v1/builds/:team/:project", ExecuteBuildHandler(k8s))
	router.GET("/api/v1/teams", TeamsHandler)
	router.GET("/api/v1/logs/:id", LogHandler(awsStorageService))
	router.GET("/api/v1/artifacts/:id", ArtifactsHandler(awsStorageService))
	router.POST("/api/v1/shutdown/:state", ShutdownHandler)
	router.POST("/hooks/:repomanager", HooksHandler(*buildScriptsRepo, *buildScriptsRepoBranch, k8s))

	var err error
	atoms, err = assembleAtoms(*buildScriptsRepo, *buildScriptsRepoBranch)
	if err != nil {
		Log.Printf("Cannot clone build scripts repository: %v\n", err)
	}
	for _, v := range atoms {
		Log.Printf("Project: %+v\n", v)
	}

	if !*noWebsocket {
		go k8s.Websock()
	}

	corsWrapper := func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Foo", "bar")
			w.Header().Set("X-Goo", "far")
			handler.ServeHTTP(w, r)
		})
	}

	Log.Println("decap ready on port 9090...")
	http.ListenAndServe(":9090", corsWrapper(router))
}
