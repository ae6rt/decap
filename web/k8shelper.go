package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"encoding/base64"
	"encoding/json"
	"io"
	"net/url"

	"github.com/ae6rt/decap/web/k8stypes"
	"github.com/ae6rt/retry"
	"github.com/pborman/uuid"
	"golang.org/x/net/websocket"
)

func NewDefaultDecap(apiServerURL, username, password, awsKey, awsSecret, awsRegion string, locker Locker, buildScriptsRepo, buildScriptsRepoBranch string) DefaultDecap {

	tlsConfig := tls.Config{}
	caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		Log.Printf("Skipping Kubernetes master TLS verify: %v\n", err)
		tlsConfig.InsecureSkipVerify = true
	} else {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
		Log.Println("Kubernetes master secured with TLS")
	}

	apiClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tlsConfig,
	}}

	data, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")

	return DefaultDecap{
		MasterURL:              apiServerURL,
		apiToken:               string(data),
		UserName:               username,
		Password:               password,
		Locker:                 locker,
		AWSAccessKeyID:         awsKey,
		AWSAccessSecret:        awsSecret,
		AWSRegion:              awsRegion,
		apiClient:              apiClient,
		maxPods:                10,
		buildScriptsRepo:       buildScriptsRepo,
		buildScriptsRepoBranch: buildScriptsRepoBranch,
	}
}

func (decap DefaultDecap) makeBaseContainer(buildEvent BuildEvent, buildID, branch string, projects map[string]Atom) k8stypes.Container {
	projectKey := buildEvent.Key()
	lockKey := decap.Locker.Key(projectKey, branch)
	return k8stypes.Container{
		Name:  "build-server",
		Image: projects[projectKey].Descriptor.Image,
		VolumeMounts: []k8stypes.VolumeMount{
			k8stypes.VolumeMount{
				Name:      "build-scripts",
				MountPath: "/home/decap/buildscripts",
			},
			k8stypes.VolumeMount{
				Name:      "decap-credentials",
				MountPath: "/etc/secrets",
			},
		},
		Env: []k8stypes.EnvVar{
			k8stypes.EnvVar{
				Name:  "BUILD_ID",
				Value: buildID,
			},
			k8stypes.EnvVar{
				Name:  "PROJECT_KEY",
				Value: projectKey,
			},
			k8stypes.EnvVar{
				Name:  "BRANCH_TO_BUILD",
				Value: branch,
			},
			k8stypes.EnvVar{
				Name:  "BUILD_LOCK_KEY",
				Value: lockKey,
			},
			k8stypes.EnvVar{
				Name:  "AWS_ACCESS_KEY_ID",
				Value: decap.AWSAccessKeyID,
			},
			k8stypes.EnvVar{
				Name:  "AWS_SECRET_ACCESS_KEY",
				Value: decap.AWSAccessSecret,
			},
			k8stypes.EnvVar{
				Name:  "AWS_DEFAULT_REGION",
				Value: decap.AWSRegion,
			},
		},
		Lifecycle: &k8stypes.Lifecycle{
			PreStop: &k8stypes.Handler{
				Exec: &k8stypes.ExecAction{
					Command: []string{
						"bctool", "unlock",
						"--lockservice-base-url", "http://lockservice.decap-system:2379",
						"--build-id", buildID,
						"--build-lock-key", lockKey,
					},
				},
			},
		},
	}
}

func (decap DefaultDecap) makeSidecarContainers(buildEvent BuildEvent, projects map[string]Atom) []k8stypes.Container {
	projectKey := buildEvent.Key()
	arr := make([]k8stypes.Container, len(projects[projectKey].Sidecars))

	for i, v := range projects[projectKey].Sidecars {
		var c k8stypes.Container
		err := json.Unmarshal([]byte(v), &c)
		if err != nil {
			Log.Println(err)
			continue
		}
		arr[i] = c
	}
	return arr
}

func (decap DefaultDecap) makePod(buildEvent BuildEvent, buildID, branch string, containers []k8stypes.Container) k8stypes.Pod {
	return k8stypes.Pod{
		TypeMeta: k8stypes.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: k8stypes.ObjectMeta{
			Name:      buildID,
			Namespace: "decap",
			Labels: map[string]string{
				"type":    "decap-build",
				"team":    buildEvent.Team(),
				"project": buildEvent.Project(),
				"branch":  branch,
			},
		},
		Spec: k8stypes.PodSpec{
			Volumes: []k8stypes.Volume{
				k8stypes.Volume{
					Name: "build-scripts",
					VolumeSource: k8stypes.VolumeSource{
						GitRepo: &k8stypes.GitRepoVolumeSource{
							Repository: decap.buildScriptsRepo,
							Revision:   decap.buildScriptsRepoBranch,
						},
					},
				},
				k8stypes.Volume{
					Name: "decap-credentials",
					VolumeSource: k8stypes.VolumeSource{
						Secret: &k8stypes.SecretVolumeSource{
							SecretName: "decap-credentials",
						},
					},
				},
			},
			Containers:    containers,
			RestartPolicy: "Never",
		},
	}
}

func (decap DefaultDecap) makeContainers(buildEvent BuildEvent, buildID, branch string, projects map[string]Atom) []k8stypes.Container {
	baseContainer := decap.makeBaseContainer(buildEvent, buildID, branch, projects)
	sidecars := decap.makeSidecarContainers(buildEvent, projects)

	containers := make([]k8stypes.Container, 0)
	containers = append(containers, baseContainer)
	containers = append(containers, sidecars...)
	return containers
}

func (decap DefaultDecap) LaunchBuild(buildEvent BuildEvent) error {
	atomKey := buildEvent.Key()

	atoms := getAtoms()

	atom := atoms[atomKey]

	for _, ref := range buildEvent.Refs() {

		if !atom.Descriptor.isRefManaged(ref) {
			Log.Printf("Ref %s is not managed on project %s.  Not launching a build.\n", ref, atomKey)
			continue
		}

		key := decap.Locker.Key(atomKey, ref)
		buildID := uuid.NewRandom().String()

		containers := decap.makeContainers(buildEvent, buildID, ref, atoms)

		pod := decap.makePod(buildEvent, buildID, ref, containers)

		podBytes, err := json.Marshal(&pod)
		if err != nil {
			Log.Println(err)
			continue
		}

		resp, err := decap.Locker.Lock(key, buildID)
		if err != nil {
			Log.Printf("Failed to acquire lock %s on build %s: %v\n", key, buildID, err)
			if err := decap.DeferBuild(buildEvent, ref); err != nil {
				Log.Printf("Failed to defer build: %+v\n", buildID)
			} else {
				Log.Printf("Deferred build: %+v\n", buildID)
			}
			continue
		}

		if resp.Node.Value == buildID {
			Log.Printf("Acquired lock on build %s with key %s\n", buildID, key)
			if podError := decap.CreatePod(podBytes); podError != nil {
				Log.Println(podError)
				if _, err := decap.Locker.Unlock(key, buildID); err != nil {
					Log.Println(err)
				} else {
					Log.Printf("Released lock on build %s with key %s because of pod creation error %v\n", buildID, key, podError)
				}
			}
			Log.Printf("Created pod=%s\n", buildID)
		} else {
			Log.Printf("Failed to acquire lock %s on build %s\n", key, buildID)
		}
	}
	return nil
}

func (decap DefaultDecap) CreatePod(pod []byte) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/namespaces/decap/pods", decap.MasterURL), bytes.NewReader(pod))
	if err != nil {
		Log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if decap.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+decap.apiToken)
	} else {
		req.SetBasicAuth(decap.UserName, decap.Password)
	}

	resp, err := decap.apiClient.Do(req)
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

func (decap DefaultDecap) DeletePod(podName string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/namespaces/decap/pods/%s", decap.MasterURL, podName), nil)
	if err != nil {
		Log.Println(err)
		return err
	}
	if decap.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+decap.apiToken)
	} else {
		req.SetBasicAuth(decap.UserName, decap.Password)
	}

	resp, err := decap.apiClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			Log.Printf("Error reading non-200 response body: %v\n", err)
			return err
		} else {
			Log.Printf("%s\n", string(data))
			return nil
		}
	}
	return nil
}

func (decap DefaultDecap) Websock() {

	var conn *websocket.Conn

	work := func() error {
		originURL, err := url.Parse(decap.MasterURL + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
		if err != nil {
			return err
		}
		serviceURL, err := url.Parse("wss://" + originURL.Host + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
		if err != nil {
			return err
		}

		var hdrs http.Header
		if decap.apiToken != "" {
			hdrs = map[string][]string{"Authorization": []string{"Bearer " + decap.apiToken}}
		} else {
			hdrs = map[string][]string{"Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte(decap.UserName+":"+decap.Password))}}
		}

		cfg := websocket.Config{
			Location:  serviceURL,
			Origin:    originURL,
			TlsConfig: &tls.Config{InsecureSkipVerify: true},
			Header:    hdrs,
			Version:   websocket.ProtocolVersionHybi13,
		}

		if conn, err = websocket.DialConfig(&cfg); err != nil {
			return err
		}
		return nil
	}

	err := retry.New(5*time.Second, 60, retry.DefaultBackoffFunc).Try(work)
	if err != nil {
		Log.Printf("Error opening websocket connection.  Will be unable to reap exited pods.: %v\n", err)
		return
	}
	Log.Print("Watching pods on websocket")

	var msg string
	for {
		err := websocket.Message.Receive(conn, &msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			Log.Println("Couldn't receive msg " + err.Error())
		}
		var pod PodWatch
		if err := json.Unmarshal([]byte(msg), &pod); err != nil {
			Log.Println(err)
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
			_, err := decap.Locker.Lock("/pods/"+pod.Object.Meta.Name, "anyvalue")
			if err == nil {
				if err := decap.DeletePod(pod.Object.Meta.Name); err != nil {
					Log.Print(err)
				} else {
					Log.Printf("Pod deleted: %s\n", pod.Object.Meta.Name)
				}
			}
		}
	}
}

func (decap DefaultDecap) DeferBuild(event BuildEvent, branch string) error {
	// only defer the most recent project/branch.  displace old deferrals.
	_ = UserBuildEvent{
		TeamFld:    event.Team(),
		ProjectFld: event.Project(),
		RefsFld:    []string{branch},
	}
	return nil
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
