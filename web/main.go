package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

var (
	apiServerBaseURL       = flag.String("api-server-base-url", "https://kubernetes", "Kubernetes API server base URL")
	apiServerUser          = flag.String("api-server-username", "admin", "Kubernetes API server username to use if no service acccount API token is present.")
	apiServerPassword      = flag.String("api-server-password", "admin123", "Kubernetes API server password to use if no service acccount API token is present.")
	buildScriptsRepo       = flag.String("build-scripts-repo", "https://github.com/ae6rt/decap-build-scripts.git", "Git repo where userland build scripts are held.")
	buildScriptsRepoBranch = flag.String("build-scripts-repo-branch", "master", "Branch or revision to use on git repo where userland build scripts are held.")
	image                  = flag.String("image", "ae6rt/decap-build-base:latest", "Build container image.")
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

	k8s := NewK8s(*apiServerBaseURL, *apiServerUser, *apiServerPassword, locker)

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/version", versionHandler)
	r.HandleFunc("/api/v1/projects", projectsHandler)
	r.HandleFunc("/api/v1/builds", projectsHandler)
	r.HandleFunc("/hooks/github", GitHubHandler{K8sBase: k8s}.handle)
	r.HandleFunc("/hooks/stash", StashHandler{K8sBase: k8s}.handle)
	r.HandleFunc("/hooks/bitbucket", BitBucketHandler{K8sBase: k8s}.handle)
	r.HandleFunc("/", documentRootHandler)
	http.Handle("/", r)

	Log.Println("decap ready on port 9090...")
	http.ListenAndServe(":9090", nil)
}
