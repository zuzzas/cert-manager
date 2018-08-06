package execute

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"github.com/jetstack/cert-manager/pkg/issuer/acme/dns/util"
	"os"
	"os/exec"
	"strings"
	"time"
)

var workingDirectory = "/srv/execute_plugin_dir/"

type DNSProvider struct {
	pluginName string
	binaryPath string
	env        []string
}

func NewDNSProvider() (*DNSProvider, error) {
	pluginName := os.Getenv("EXECUTE_PLUGIN_NAME")
	env := envStrToSlice(os.Getenv("EXECUTE_ENV"))

	return NewDNSProviderCredentials(pluginName, env)
}

func NewDNSProviderCredentials(pluginName string, env []string) (*DNSProvider, error) {
	if pluginName == "" || len(env) == 0 {
		return nil, fmt.Errorf("execute parameters missing")
	}

	return &DNSProvider{
		pluginName: pluginName,
		binaryPath: workingDirectory + pluginName,
		env:        env,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (c *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 2 * time.Second
}

func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := util.DNS01Record(domain, keyAuth)
	args := []string{"present", fqdn, value}

	err, _, _ := execute(c.binaryPath, args, c.env)

	return err
}

func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := util.DNS01Record(domain, keyAuth)
	args := []string{"cleanup", fqdn, ""}

	err, _, _ := execute(c.binaryPath, args, c.env)

	return err
}

func execute(binaryPath string, args, env []string) (error, string, string) {
	var stdoutBuf, stderrBuf bytes.Buffer
	var exitCode uint

	cmd := exec.Cmd{
		Path:   binaryPath,
		Args:   append([]string{binaryPath}, args...),
		Env:    env,
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
	}

	glog.Infof("executing program %v with the following arguments: %v. Environment: %v", cmd.Path, cmd.Args, cmd.Env)
	status := cmd.Run()

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	glog.Infof("process finished with exit code: %v, stdout: %v, stderr: %v",
		exitCode, stdout, stderr)

	return status, stdout, stderr
}

func envStrToSlice(str string) []string {
	var envSlice []string

	for _, v := range strings.Split(str, ",") {
		envSlice = append(envSlice, strings.TrimSpace(v))
	}

	return envSlice
}
