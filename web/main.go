package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ae6rt/decap/web/credentials"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/lock"
	"github.com/ae6rt/decap/web/scmclients"
	"github.com/ae6rt/decap/web/storage"
	"github.com/julienschmidt/httprouter"
)

var (
	awsKey                 = flag.String("aws-access-key", "", "Default decap AWS access key.  /etc/secrets/aws-key in the cluster overrides this.")
	awsSecret              = flag.String("aws-secret-key", "", "Default decap AWS access secret.  /etc/secrets/aws-secret in the cluster overrides this.")
	awsRegion              = flag.String("aws-region", "us-west-1", "Default decap AWS region.  /etc/secrets/aws-region in the cluster overrides this.")
	githubClientID         = flag.String("github-client-id", "", "Default Github ClientID for quering Github repos.  /etc/secrets/github-client-id in the cluster overrides this.")
	githubClientSecret     = flag.String("github-client-secret", "", "Default Github Client Secret for quering Github repos.  /etc/secrets/github-client-secret in the cluster overrides this.")
	buildScriptsRepo       = flag.String("build-scripts-repo", "https://github.com/ae6rt/decap-build-scripts.git", "Git repo where userland build scripts are held.")
	buildScriptsRepoBranch = flag.String("build-scripts-repo-branch", "master", "Branch or revision to use on git repo where userland build scripts are held.")
	noPodWatcher           = flag.Bool("no-podwatcher", false, "Do not start k8s podwatcher, for dev mode purposes.")
	versionFlag            = flag.Bool("version", false, "Print version info and exit.")
)

var (
	// Log is the global logger.
	Log          = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	buildVersion string
	buildCommit  string
	buildDate    string
	buildGoSDK   string
)

func main() {
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

	// Create collaborating services.

	k8sClient, err := NewKubernetesClient()
	if err != nil {
		Log.Fatalf("Cannot create Kubernetes client: %v\n", err)
	}
	deferralService := deferrals.NewDefault(Log)

	lockService := lock.NewDefault(k8sClient)

	projectManager := NewDefaultProjectManager(BuildScripts{URL: *buildScriptsRepo, Branch: *buildScriptsRepoBranch})

	buildManager := NewBuildManager(k8sClient, projectManager, lockService, deferralService, Log)

	buildStore := storage.NewAWS(credentials.AWSCredential{AccessKey: *awsKey, AccessSecret: *awsSecret, Region: *awsRegion}, Log)

	scmManagers := map[string]scmclients.SCMClient{
		"github": scmclients.NewGithub("https://api.github.com", *githubClientID, *githubClientSecret),
	}

	// Setup HTTP routing

	router := httprouter.New()

	// This serves up the front end webapp.
	router.ServeFiles("/decap/*filepath", http.Dir("./static"))

	// Set options for frontend webapp.
	router.OPTIONS("/api/v1/*filepath", HandleOptions)

	// Report backend version.
	router.GET("/api/v1/version", VersionHandler)

	// Report managed projects
	router.GET("/api/v1/projects", ProjectsHandler)

	// Report on branches of a given project
	router.GET("/api/v1/projects/:team/:project/refs", ProjectRefsHandler(scmManagers))

	// Report on historical builds for a given project
	router.GET("/api/v1/builds/:team/:project", BuildsHandler(buildStore))

	// Terminates a running build
	router.DELETE("/api/v1/builds/:id", StopBuildHandler(buildManager, Log))

	// Execute a build on demand
	router.POST("/api/v1/builds/:team/:project", ExecuteBuildHandler(buildManager, Log))

	// Report on teams.  (I think this is a project - msp)
	router.GET("/api/v1/teams", TeamsHandler)

	// Report on currenly deferred builds.
	router.GET("/api/v1/deferred", DeferredBuildsHandler(buildManager))

	// Remove a build from the deferred builds queue.
	router.POST("/api/v1/deferred", DeferredBuildsHandler(buildManager))

	//	Return gzipped console log, or console log in plain text if Accept: text/plain is set
	router.GET("/api/v1/logs/:id", LogHandler(buildStore))

	// ArtifactsHandler returns build artifacts gzipped tarball, or file listing in tarball if Accept: text/plain is set
	router.GET("/api/v1/artifacts/:id", ArtifactsHandler(buildStore))

	// Return current state of the build queue
	router.GET("/api/v1/shutdown", ShutdownHandler(buildManager))

	// ShutdownHandler stops the build queue from accepting new build requests.
	router.POST("/api/v1/shutdown/:state", ShutdownHandler(buildManager))

	// LogLevelHandler toggles debug logging.
	router.POST("/api/v1/loglevel/:level", LogLevelHandler)

	// The interface for external SCM systems to post VCS events through post-commit hooks.
	router.POST("/hooks/:repomanager", HooksHandler(projectManager, buildManager, Log))

	projects, err := projectManager.Assemble()
	if err != nil {
		Log.Printf("Cannot clone build scripts repository: %v\n", err)
	}

	go func() {
		c := time.Tick(1 * time.Minute)
		buildManager.LaunchDeferred(c)
	}()
	go projectMux(projects)
	go logLevelMux(LogDefault)
	go shutdownMux(BuildQueueOpen)
	if !*noPodWatcher {
		go buildManager.PodWatcher()
	}

	Log.Println("decap ready on port 9090...")
	Log.Fatal(http.ListenAndServe(":9090", corsWrapper(router)))
}
