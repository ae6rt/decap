package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

var (
	apiServerBaseURL       = flag.String("api-server-base-url", "https://kubernetes", "Kubernetes API server base URL")
	apiServerUser          = flag.String("api-server-username", "admin", "Kubernetes API server username to use if no service acccount API token is present.")
	apiServerPassword      = flag.String("api-server-password", "admin123", "Kubernetes API server password to use if no service acccount API token is present.")
	awsAccessKey           = flag.String("aws-access-key", "", "Default decap AWS access key.  /etc/secrets/aws-key in the cluster overrides this.")
	awsSecret              = flag.String("aws-secret-key", "", "Default decap AWS access secret.  /etc/secrets/aws-secret in the cluster overrides this.")
	awsRegion              = flag.String("aws-region", "us-west-1", "Default decap AWS region.  /etc/secrets/aws-region in the cluster overrides this.")
	buildScriptsRepo       = flag.String("build-scripts-repo", "https://github.com/ae6rt/decap-build-scripts.git", "Git repo where userland build scripts are held.")
	buildScriptsRepoBranch = flag.String("build-scripts-repo-branch", "master", "Branch or revision to use on git repo where userland build scripts are held.")
	image                  = flag.String("image", "ae6rt/decap-build-base:latest", "Build container image that runs userland build scripts.")
	versionFlag            = flag.Bool("version", false, "Print version info and exit.")

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
}

func main() {
	locker := NewDefaultLock([]string{"http://localhost:2379"})
	k8s := NewDefaultDecap(*apiServerBaseURL, *apiServerUser, *apiServerPassword, locker)
	awsStorageService := NewAWSStorageService(*awsAccessKey, *awsSecret, *awsRegion)

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/version", VersionHandler)
	router.GET("/builds", BuildsHandler(awsStorageService))
	router.GET("/builds/:id/logs", LogHandler(awsStorageService))
	router.GET("/builds/:id/artifacts", ArtifactsHandler(awsStorageService))
	router.POST("/hooks/:repomanager", HooksHandler(k8s))

	Log.Println("decap ready on port 9090...")
	http.ListenAndServe(":9090", router)
}
