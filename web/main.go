package main

import (
	"flag"
	"io/ioutil"
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
	versionFlag            = flag.Bool("version", false, "Print version info and exit.")

	repoManagerClientCredentials = make(map[string]RepoManagerCredential)

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

	repoManagerClientCredentials["github"] = RepoManagerCredential{User: *githubClientID, Password: *githubClientSecret}
}

func main() {
	locker := NewDefaultLock([]string{"http://localhost:2379"})
	k8s := NewDefaultDecap(*apiServerBaseURL, *apiServerUser, *apiServerPassword, *awsKey, *awsSecret, *awsRegion, locker)
	awsStorageService := NewAWSStorageService(*awsKey, *awsSecret, *awsRegion)

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/api/v1/version", VersionHandler)
	router.GET("/api/v1/projects", ProjectsHandler)
	router.GET("/api/v1/projects/:team/:library/branches", ProjectBranchesHandler(repoManagerClientCredentials))
	router.GET("/api/v1/builds/:team/:library", BuildsHandler(awsStorageService))
	router.DELETE("/api/v1/builds/:id", StopBuildHandler(k8s))
	router.POST("/api/v1/builds/:team/:library", ExecuteBuildHandler(k8s))
	router.GET("/api/v1/logs/:id", LogHandler(awsStorageService))
	router.GET("/api/v1/teams", TeamsHandler)
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
