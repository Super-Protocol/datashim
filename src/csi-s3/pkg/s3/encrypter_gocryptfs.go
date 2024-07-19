package s3

import (
	"fmt"
	"os/exec"
	"path/filepath"
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
	tempInitDirPath := filepath.Join(targetDir, "temp-init")

	CreateFolderIfNotExists(tempInitDirPath)

	// delete tempInitDirPath with all the content in any case
	defer func() {
		exec.Command("rm", "-rf", tempInitDirPath)
	}()

	args := []string{
		"-init",
		"-passfile", passFile,
		tempInitDirPath,
	}
	cmd := exec.Command(gocryptfsCmd, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error gocryptfs initialize command: %s\nargs: %s\noutput: %s", gocryptfsCmd, args, out)
	}

	copyCommandArgs := []string{
		tempInitDirPath + "/*",
		target,
	}
	cmd = exec.Command("cp", copyCommandArgs...)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error gocryptfs copy to target path on init, nargs: %s\noutput: %s", args, out)
	}

	return nil
}
