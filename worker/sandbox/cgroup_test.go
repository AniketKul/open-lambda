package sandbox

import (
	"bytes"
	"fmt"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/open-lambda/open-lambda/worker/config"
	"github.com/open-lambda/open-lambda/worker/dockerutil"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"testing"
	"time"
)

var testDir string
var baseDir string
var conf *config.Config

func Init() {
	var err error

	// docker client
	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal("failed to get docker client: ", err)
	}

	// test dir, base dir
	testDir, err = ioutil.TempDir(os.TempDir(), "cgroup_test")
	if err != nil {
		log.Fatal("cannot create temp dir")
	}
	baseDir = path.Join(testDir, "base")
	fmt.Printf("Using %s for cgroup testing\n", testDir)

	// conf
	conf, err = config.ParseConfig(os.Getenv("WORKER_CONFIG"))
	if err != nil {
		log.Fatal(err)
	}
	conf.Cgroup_base = baseDir

	// ubuntu FS base
	fmt.Printf("dump lambda root to %s\n", baseDir)
	err = dockerutil.DumpDockerImage(client, "lambda", baseDir)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	Init()
	res := m.Run()
	// TODO: cleanup
	os.Exit(res)
}

func TestCreate(t *testing.T) {
	factory, err := NewCgroupSBFactory(conf)
	if err != nil {
		t.Fatal(err.Error())
	}

	handler_name := "hello"

	handler_dir := path.Join(conf.Reg_dir, handler_name)
	sandbox_dir := path.Join(testDir, "sandbox1")
	if err := os.Mkdir(sandbox_dir, 0777); err != nil {
		t.Fatal(err.Error())
	}

	s, err := factory.Create(handler_dir, sandbox_dir)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err := s.Start(); err != nil {
		t.Fatal(err.Error())
	}

	channel, err := s.Channel()
	if err != nil {
		t.Fatal(err.Error())
	}

	time.Sleep(1000 * time.Millisecond)

	// forward request
	url := fmt.Sprintf("http://container/runLambda/%s", handler_name)
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Transport: &channel.Transport}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err.Error())
	}
	wbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err.Error())
	}
	if string(wbody) != "\"hello\"" {
		t.Fatal(fmt.Sprintf("Unexpected resp: '%s'", string(wbody)))
	}
}
