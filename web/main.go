package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/distrlocks"
	"github.com/ae6rt/decap/web/scmclients"
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

	// todo this is a debugging aid and should be handled not so heavy-handed
	noWebsocket = flag.Bool("no-websocket", false, "Do not start websocket client that watches pods.")

	versionFlag = flag.Bool("version", false, "Print version info and exit.")

	// Log is the logger for package main
	Log = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

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

	distributedLocker := distrlocks.NewDynamoDbLockService(distrlocks.NewDynamoDB(*awsKey, *awsSecret, *awsRegion))

	deferralService := deferrals.NewInMemoryDeferralService(Log)

	buildLauncher := NewBuilder(*apiServerBaseURL, *apiServerUser, *apiServerPassword, *awsKey, *awsSecret, *awsRegion, *buildScriptsRepo, *buildScriptsRepoBranch, distributedLocker, deferralService, Log)
	storageService := NewAWSStorageService(*awsKey, *awsSecret, *awsRegion)
	scmManagers := map[string]scmclients.SCMClient{
		"github": scmclients.NewGithubClient("https://api.github.com", *githubClientID, *githubClientSecret),
	}

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
	router.GET("/api/v1/builds/:team/:project", BuildsHandler(storageService))

	// Terminates a running build
	router.DELETE("/api/v1/builds/:id", StopBuildHandler(buildLauncher))

	// Execute a build on demand
	router.POST("/api/v1/builds/:team/:project", ExecuteBuildHandler(buildLauncher))

	// Report on teams.  (I think this is a project - msp)
	router.GET("/api/v1/teams", TeamsHandler)

	// Report on currenly deferred builds.
	router.GET("/api/v1/deferred", DeferredBuildsHandler(buildLauncher))

	// Remove a build from the deferred builds queue.
	router.POST("/api/v1/deferred", DeferredBuildsHandler(buildLauncher))

	//	Return gzipped console log, or console log in plain text if Accept: text/plain is set
	router.GET("/api/v1/logs/:id", LogHandler(storageService))

	// ArtifactsHandler returns build artifacts gzipped tarball, or file listing in tarball if Accept: text/plain is set
	router.GET("/api/v1/artifacts/:id", ArtifactsHandler(storageService))

	// Return current state of the build queue
	router.GET("/api/v1/shutdown", ShutdownHandler)

	// ShutdownHandler stops the build queue from accepting new build requests.
	router.POST("/api/v1/shutdown/:state", ShutdownHandler)

	// LogLevelHandler toggles debug logging.
	router.POST("/api/v1/loglevel/:level", LogLevelHandler)

	// The interface for external SCM systems to post VCS events through post-commit hooks.
	router.POST("/hooks/:repomanager", HooksHandler(*buildScriptsRepo, *buildScriptsRepoBranch, buildLauncher))

	projects, err := assembleProjects(*buildScriptsRepo, *buildScriptsRepoBranch)
	if err != nil {
		Log.Printf("Cannot clone build scripts repository: %v\n", err)
	}

	// Wrap the deferred launcher like this so we can more easily test LaunchDeferred(c) method.
	go func() {
		c := time.Tick(1 * time.Minute)
		buildLauncher.LaunchDeferred(c)
	}()
	go projectMux(projects)
	go logLevelMux(LogDefault)
	go shutdownMux(BuildQueueOpen)
	if !*noWebsocket {
		go buildLauncher.PodWatcher()
	}

	Log.Println("decap ready on port 9090...")
	Log.Fatal(http.ListenAndServe(":9090", corsWrapper(router)))
}
