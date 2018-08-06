package execute

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

const scriptContents = `#!/bin/sh

set -e

unset PWD

case "$1" in
    present)
        echo "$*"
        echo "$(env)" 1>&2
        ;;
    cleanup)
        echo "$*"
        echo "$(env)" 1>&2
        ;;
    *)
        echo "argument $1 is not recognized"
        exit 1
esac
`

var (
	execBinaryPath string
	execEnvStr     string
	execEnv        []string
)

func TestMain(m *testing.M) {
	execEnvStr = "TEST=indeed,VICTORY=ahead"
	execEnv = envStrToSlice(execEnvStr)

	script, err := ioutil.TempFile("", "cert-manager-exec-test-")
	if err != nil {
		log.Fatalln(err)
	}
	execBinaryPath = script.Name()

	if script.Chmod(0700); err != nil {
		log.Fatalln(err)
	}

	_, err = script.WriteString(scriptContents)
	if err != nil {
		log.Fatalln(err)
	}

	if script.Close(); err != nil {
		log.Println(err)
	}

	status := m.Run()

	if os.Remove(script.Name()); err != nil {
		log.Println(err)
	}

	os.Exit(status)
}

func restoreExecuteEnv() {
	os.Setenv("EXECUTE_PLUGIN_NAME", execBinaryPath)
	os.Setenv("EXECUTE_ENV", execEnvStr)
}

func TestNewDNSProviderValid(t *testing.T) {
	os.Setenv("EXECUTE_PLUGIN_NAME", "")
	os.Setenv("EXECUTE_ENV", "")
	_, err := NewDNSProviderCredentials("123", []string{"WATER=bread", "LAVA=charcoal"})
	assert.NoError(t, err)
	restoreExecuteEnv()
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	os.Setenv("EXECUTE_PLUGIN_NAME", "regru-issuer")
	os.Setenv("EXECUTE_ENV", "REGRU_USER=test@example.com,REGRU_PASS=123")
	_, err := NewDNSProvider()
	assert.NoError(t, err)
	restoreExecuteEnv()
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	os.Setenv("EXECUTE_PLUGIN_NAME", "")
	os.Setenv("EXECUTE_ENV", "")
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "execute parameters missing")
	restoreExecuteEnv()
}

func TestExecutePresent(t *testing.T) {
	provider, err := NewDNSProviderCredentials(execBinaryPath, []string{"WATER=bread", "LAVA=charcoal"})
	provider.binaryPath = execBinaryPath
	assert.NoError(t, err)

	err = provider.Present("132fds.example.com", "", "123dda=")
	assert.NoError(t, err)
}

func TestExecuteCleanUp(t *testing.T) {
	provider, err := NewDNSProviderCredentials(execBinaryPath, []string{"WATER=bread", "LAVA=charcoal"})
	provider.binaryPath = execBinaryPath
	assert.NoError(t, err)

	err = provider.CleanUp("132fds.example.com", "", "123dda=")
	assert.NoError(t, err)
}

func TestExecutePresentExecution(t *testing.T) {
	expectedStdout := "present 12.example.com testasd=\n"
	expectedStderr := "VICTORY=ahead\nTEST=indeed\n"

	err, stdout, stderr := execute(execBinaryPath, []string{"present", "12.example.com", "testasd="}, execEnv)
	assert.NoError(t, err)
	assert.Equal(t, expectedStdout, stdout)
	assert.Equal(t, expectedStderr, stderr)
}

func TestExecuteCleanUpExecution(t *testing.T) {
	expectedStdout := "cleanup 12.example.com testasd=\n"
	expectedStderr := "VICTORY=ahead\nTEST=indeed\n"

	err, stdout, stderr := execute(execBinaryPath, []string{"cleanup", "12.example.com", "testasd="}, execEnv)
	assert.NoError(t, err)
	assert.Equal(t, expectedStdout, stdout)
	assert.Equal(t, expectedStderr, stderr)
}
