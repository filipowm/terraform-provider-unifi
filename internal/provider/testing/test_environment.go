package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

const (
	TestEnvStarting TestEnvironmentStatus = iota
	TestEnvReady
	TestEnvDown
	TestEnvUnknown
)

var mutex = sync.Mutex{}

type TestEnvironmentStatus int

type TestEnvironment struct {
	Client         unifi.Client
	Endpoint       string
	Shutdown       func()
	ctx            context.Context
	internalClient *http.Client
	mutex          sync.Mutex
	timeout        time.Duration
}

var testEnv *TestEnvironment
var testClient unifi.Client

type envStatus struct {
	Meta struct {
		Up bool `json:"up"`
	} `json:"meta"`
}

func Run(m *testing.M) int {
	if os.Getenv("TF_ACC") == "" {
		// short circuit non-acceptance test runs
		os.Exit(m.Run())
	}
	env := newTestEnvironment(5 * time.Minute)
	return env.Run(m)
}

func newTestEnvironment(startupTimeout time.Duration) *TestEnvironment {
	mutex.Lock()
	defer mutex.Unlock()
	if testEnv != nil {
		return testEnv
	}
	c := http.Client{Transport: provider.CreateHttpTransport(true)}
	ctx := context.Background()
	testEnv = &TestEnvironment{
		Endpoint:       "https://localhost:8443", // default endpoint, assumed
		timeout:        startupTimeout,
		mutex:          sync.Mutex{},
		ctx:            ctx,
		internalClient: &c,
		Shutdown:       func() {},
	}
	return testEnv
}

func (te *TestEnvironment) IsReady() bool {
	if st, _ := te.readStatus(te.ctx); st != TestEnvReady {
		return false
	}
	return true
}

func (te *TestEnvironment) Run(m *testing.M) int {
	mutex.Lock() // run one by one
	defer mutex.Unlock()
	err := te.start()
	defer func() {
		te.Shutdown()
	}()
	if err != nil {
		panic(err)
	}
	err = te.waitUntilReady()
	if err != nil {
		panic(err)
	}
	c, err := te.NewTestClient()
	if err != nil {
		panic(err)
	}
	te.Client = c
	return m.Run()
}

func (te *TestEnvironment) readStatus(ctx context.Context) (TestEnvironmentStatus, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/status", te.Endpoint), nil)
	if err != nil {
		return TestEnvUnknown, err
	}
	req = req.WithContext(ctx)
	r, err := te.internalClient.Do(req)
	if err != nil {
		return TestEnvDown, err
	}
	resp := envStatus{}
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return TestEnvUnknown, err
	}
	if resp.Meta.Up {
		return TestEnvReady, nil
	}
	return TestEnvStarting, nil
}

func findFileInProject(filename string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Walk up the directory tree until we find a file
	for {
		path := filepath.Join(wd, filename)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		if wd == "/" {
			break
		}
		wd = filepath.Dir(wd)
	}
	return "", fmt.Errorf("file %s not found in project", filename)
}

func (te *TestEnvironment) startDockerController(ctx context.Context) error {
	composeFile, err := findFileInProject("docker-compose.yaml")
	dc, err := compose.NewDockerCompose(composeFile)
	shutdown := func() {
		if dc != nil {
			if err := dc.Down(context.Background(), compose.RemoveOrphans(true), compose.RemoveImagesLocal); err != nil {
				panic(err)
			}
		}
	}
	te.Shutdown = shutdown
	if err != nil {
		return err
	}

	if err = dc.WithOsEnv().Up(ctx, compose.Wait(true)); err != nil {
		return fmt.Errorf("failed to start docker-compose. Controller container might be already running or starting: %w", err)
	}
	container, err := dc.ServiceContainer(ctx, "unifi")

	// Dump the container logs on exit.
	//
	// TODO: Use https://pkg.go.dev/github.com/testcontainers/testcontainers-go#LogConsumer instead.
	te.Shutdown = func() {
		shutdown()

		if os.Getenv("UNIFI_STDOUT") == "" {
			return
		}

		stream, err := container.Logs(ctx)
		if err != nil {
			panic(err)
		}

		buffer := new(bytes.Buffer)
		buffer.ReadFrom(stream)
		testcontainers.Logger.Printf("%s", buffer)
	}
	endpoint, err := container.PortEndpoint(ctx, "8443/tcp", "https")
	if err != nil {
		return err
	}
	te.Endpoint = endpoint
	return nil
}

func (te *TestEnvironment) waitUntilReady() error {
	te.mutex.Lock()
	ctx, cancel := context.WithTimeoutCause(te.ctx, te.timeout, fmt.Errorf("controller was not ready within %s", te.timeout))
	defer cancel()
	defer te.mutex.Unlock()
	if st, _ := te.readStatus(ctx); st == TestEnvDown || st == TestEnvUnknown {
		return fmt.Errorf("controller is not starting nor running. Use start() first to start the controller")
	}
	te.waitForController(ctx)
	if !te.IsReady() {
		return fmt.Errorf("controller is not ready within %s", te.timeout)
	}
	return nil
}

func (te *TestEnvironment) start() error {
	tflog.Error(te.ctx, "Starting test environment")
	if te.IsReady() {
		tflog.Warn(te.ctx, "Environment is already running at "+te.Endpoint)
		if te.Client == nil {
			c, err := te.NewTestClient()
			if err != nil {
				return err
			}
			te.Client = c
		}
		return nil
	}
	ctx, cancel := context.WithTimeoutCause(te.ctx, te.timeout, fmt.Errorf("controller did not start within %s", te.timeout))
	defer cancel()
	err := te.startDockerController(ctx)
	if err != nil {
		return err
	}
	tflog.Info(te.ctx, "Environment is starting at "+te.Endpoint)
	return nil
}

func (te *TestEnvironment) waitForController(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			if st, err := te.readStatus(ctx); err != nil {
				return
			} else if st == TestEnvReady {
				wg.Done()
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()
	wg.Wait()
}

func TestClient() unifi.Client {
	if testClient == nil {
		panic("Test client is not initialized")
	}
	return testClient
}

func (te *TestEnvironment) NewTestClient() (unifi.Client, error) {
	const user = "admin"
	const password = "admin"
	var err error
	if err = os.Setenv("UNIFI_USERNAME", user); err != nil {
		return nil, err
	}

	if err = os.Setenv("UNIFI_PASSWORD", password); err != nil {
		return nil, err
	}

	if err = os.Setenv("UNIFI_INSECURE", "true"); err != nil {
		return nil, err
	}

	if err = os.Setenv("UNIFI_API", te.Endpoint); err != nil {
		return nil, err
	}

	c, err := unifi.NewClient(&unifi.ClientConfig{
		URL:      te.Endpoint,
		User:     user,
		Password: password,
		HttpRoundTripperProvider: func() http.RoundTripper {
			return provider.CreateHttpTransport(true)
		},
		ValidationMode: unifi.DisableValidation,
		Logger:         unifi.NewDefaultLogger(unifi.WarnLevel),
	})
	if err == nil {
		testClient = c
	}
	return c, err
}
