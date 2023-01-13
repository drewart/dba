package cli

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"

	"github.com/drewart/dba/config"
)

func OpenMyCLIMany(configs []config.Config, command string) {
	if command == "" {
		command = "mysql"
	}

	cmdFileFmt := `#!/bin/bash 
echo "connecting to %s"
sleep 1
# 
xpanes -l ev -c "%s --defaults-file={}" %s 
echo command exited press enter to exit terminal
read varname
`
	configHosts := ""
	configFiles := ""
	for _, config := range configs {
		configHosts += config.Host + " "
		configFiles += config.FilePath + " "
	}

	fileTxt := fmt.Sprintf(cmdFileFmt, configHosts, command, configFiles)
	temp := os.TempDir()
	timeStr := time.Now().Format("2006-01-02-15-04-05")
	cmdFile := path.Join(temp, "dba-"+timeStr+".sh")

	ioutil.WriteFile(cmdFile, []byte(fileTxt), 0o777)

	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("gnome-terminal", "--command", cmdFile).Start()
	// case "windows":
	//	err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", "-a", "terminal", cmdFile).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Println(err)
	}
}

func OpenMyCLI(cnf *config.Config, command string) {
	if command == "" {
		command = "mysql"
	}

	cmdFileFmt := `#!/bin/bash 
echo "connecting to %s"
# 
%s --defaults-file=%s
echo command exited press enter to exit terminal
read varname
`

	fileTxt := fmt.Sprintf(cmdFileFmt, cnf.Host, command, cnf.FilePath)
	temp := os.TempDir()
	timeStr := time.Now().Format("2006-01-02-15-04-05")
	cmdFile := path.Join(temp, "dba-"+timeStr+".sh")

	ioutil.WriteFile(cmdFile, []byte(fileTxt), 0o777)

	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("gnome-terminal", "--command", cmdFile).Start()
	// case "windows":
	//	err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", "-a", "terminal", cmdFile).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Println(err)
	}
}

// OPenCLIInShell
func OpenCLIInShell(cnf *config.Config, command string) {
	if command == "" {
		command = "mysql"
	}

	fmt.Printf("running: %s --defaults-file=%s", command, cnf.FilePath)
	cmd := exec.Command(command, fmt.Sprintf("--defaults-file=%s", cnf.FilePath))
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}
