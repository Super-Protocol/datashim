package s3

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/golang/glog"
)

type gocryptfsEncrypter struct{}

func (enc *gocryptfsEncrypter) MountEncrypt(source string, target string, pass string) error {
	targetDir := filepath.Dir(target)
	passFile := filepath.Join(targetDir, "pass")

	err := CreateTextFile(passFile, pass)
	if err != nil {
		return err
	}

	defer DeleteFile(passFile)

	args := []string{
		"-passfile", passFile,
		source,
		target,
	}

	configFile := filepath.Join(source, "gocryptfs.conf")
	if !FileExists(configFile) {
		err := enc.initialize(source, passFile)
		if err != nil {
			return err
		}
	}

	err = fuseMount(target, gocryptfsCmd, args)
	if err != nil {
		return err
	}

	return nil
}

func (enc *gocryptfsEncrypter) initialize(target string, passFile string) error {
	targetDir := filepath.Dir(target)
	tempInitDirName := "temp-init"
	tempInitDirPath := filepath.Join(targetDir, tempInitDirName)

	CreateFolderIfNotExists(tempInitDirPath)

	// delete tempInitDirPath with all the content in any case
	defer func() {
		cmd := exec.Command("rm", "-rf", tempInitDirPath)
		_, err := cmd.CombinedOutput()

		if err != nil {
			glog.V(2).Infof("error on delete temp init folder %s, err: %s", tempInitDirPath, err)
		}
	}()

	args := []string{
		"-init",
		"-passfile", passFile,
		tempInitDirPath,
		"--debug",
		"--nosyslog",
	}
	cmd := exec.Command(gocryptfsCmd, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error gocryptfs initialize command: %s\nargs: %s\noutput: %s", gocryptfsCmd, args, out)
	}

	copyCommandArgs := []string{
		"-c",
		fmt.Sprintf("cp %s/* %s", tempInitDirPath, target),
	}
	cmd = exec.Command("/bin/sh", copyCommandArgs...)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error gocryptfs copy to target path on init, nargs: %s\noutput: %s", args, out)
	}

	return nil
}
