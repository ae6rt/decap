package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	stdlog "log"

	"github.com/ae6rt/decap/web/api/v1"
	"github.com/ae6rt/decap/web/app"
	"github.com/ae6rt/decap/web/cluster"
	"github.com/ae6rt/decap/web/credentials"
	"github.com/ae6rt/decap/web/deferrals"
	"github.com/ae6rt/decap/web/lock"
	"github.com/ae6rt/decap/web/projects"
	"github.com/ae6rt/decap/web/scmclients"
	"github.com/ae6rt/decap/web/storage"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log"
)

var (
	buildVersion string
	buildCommit  string
	buildDate    string
	buildGoSDK   string
)

func main() {

	var (
		awsKey                 = flag.String("aws-access-key", "", "Default decap AWS access key.  /etc/secrets/aws-key in the cluster overrides this.")
		awsSecret              = flag.String("aws-secret-key", "", "Default decap AWS access secret.  /etc/secrets/aws-secret in the cluster overrides this.")
		awsRegion              = flag.String("aws-region", "us-west-1", "Default decap AWS region.  /etc/secrets/aws-region in the cluster overrides this.")
		githubClientID         = flag.String("github-client-id", "", "Default Github ClientID for quering Github repos.  /etc/secrets/github-client-id in the cluster overrides this.")
		githubClientSecret     = flag.String("github-client-secret", "", "Default Github Client Secret for quering Github repos.  /etc/secrets/github-client-secret in the cluster overrides this.")
		buildScriptsRepo       = flag.String("build-scripts-repo", "https://github.com/ae6rt/decap-build-scripts.git", "Git repo where userland build scripts are held.")
		buildScriptsRepoBranch = flag.String("build-scripts-repo-branch", "master", "Branch or revision to use on git repo where userland build scripts are held.")
		noPodWatcher           = flag.Bool("no-podwatcher", false, "Do not start k8s podwatcher, for dev mode purposes.")
		httpAddr               = flag.String("http.addr", ":8080", "HTTP listen address")
		versionFlag            = flag.Bool("version", false, "Print version info and exit.")
	)
	flag.Parse()
	fmt.Println(noPodWatcher)

	fmt.Printf("Version: %s, Commit: %s, Date: %s, Go SDK: %s\n", buildVersion, buildCommit, buildDate, buildGoSDK)
	if *versionFlag {
		os.Exit(0)
	}

	*awsKey = cluster.KubeSecret("/etc/secrets/aws-key", *awsKey)
	*awsSecret = cluster.KubeSecret("/etc/secrets/aws-secret", *awsSecret)
	*awsRegion = cluster.KubeSecret("/etc/secrets/aws-region", *awsRegion)
	*githubClientID = cluster.KubeSecret("/etc/secrets/github-client-id", *githubClientID)
	*githubClientSecret = cluster.KubeSecret("/etc/secrets/github-client-secret", *githubClientSecret)

	k8sClient, err := cluster.NewKubernetesClient()
	if err != nil {
		stdlog.Fatalf("Cannot create Kubernetes client: %v\n", err)
	}

	logger := kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stdout))
	{
		logger = kitlog.NewLogfmtLogger(os.Stderr)
		logger = kitlog.With(logger, "ts", log.DefaultTimestampUTC)
		logger = kitlog.With(logger, "caller", log.DefaultCaller)
	}
	stdlog.SetOutput(kitlog.NewStdlibAdapter(logger))
	logme := stdlog.New(kitlog.NewStdlibAdapter(logger), "", stdlog.Ldate|stdlog.Ltime|stdlog.Lshortfile)

	deferralService := deferrals.NewDefault(logme)
	buildStore := storage.NewAWS(credentials.AWSCredential{AccessKey: *awsKey, AccessSecret: *awsSecret, Region: *awsRegion}, logme)
	lockService := lock.NewDefault(k8sClient)

	projectManager := projects.NewDefaultManager(*buildScriptsRepo, *buildScriptsRepoBranch, logme)

	scmManagers := map[string]scmclients.SCMClient{
		"github": scmclients.NewGithub("https://api.github.com", *githubClientID, *githubClientSecret),
	}

	s := app.New(v1.Version{}, k8sClient, deferralService, buildStore, lockService, projectManager, scmManagers)
	fmt.Println(s)

	var h http.Handler
	{
		h = app.MakeHTTPHandler(s, kitlog.With(logger, "component", "HTTP"))
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		_ = logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, h)
	}()

	_ = logger.Log("exit", <-errs)
}
