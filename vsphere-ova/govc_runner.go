package vsphere_ova

import (
	"bytes"
	"os/exec"
	"io"
	"fmt"
	"io/ioutil"
)

type govcRunner struct {
	InsecureFlag  bool
	User          string
	Password      string
	VSphereServer string
}

func (r *govcRunner) Execute(args []string) (*bytes.Buffer, *bytes.Buffer, error) {
	var outBufWriter bytes.Buffer
	var errBufWriter bytes.Buffer

	outWriter := ioutil.Discard
	errWriter := ioutil.Discard

	command := exec.Command("govc", args...)
	command.Env = r.appendEnvs()

	command.Stdout = io.MultiWriter(outWriter, &outBufWriter)
	command.Stderr = io.MultiWriter(errWriter, &errBufWriter)

	return &outBufWriter, &errBufWriter, command.Run()
}

func (r *govcRunner) appendEnvs() []string {
	var insecure = "0"
	if r.InsecureFlag {
		insecure = "1"
	}

	return []string{
		fmt.Sprintf("GOVC_URL=%s", r.VSphereServer),
		fmt.Sprintf("GOVC_USERNAME=%s", r.User),
		fmt.Sprintf("GOVC_PASSWORD=%s", r.Password),
		fmt.Sprintf("GOVC_INSECURE=%s", insecure),
	}
}
