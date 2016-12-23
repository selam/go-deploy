package main

import (
	"os"
	"strings"
	"fmt"
	"path"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
	"runtime"
	"os/exec"
	"bufio"
	"time"
)

var (
	DOT_DIRS = []string{".git"}
)


func findApplicationInPath(pth, app string) ([]string) {
	paths := make([]string,0, 0)
	filepath.Walk(pth, func(p string, info os.FileInfo, err error) error{
		if !info.IsDir() {
			return nil
		}

		for _, ignore := range DOT_DIRS {
			if strings.HasSuffix(p, ignore) {
				return filepath.SkipDir
			}
		}

		if strings.HasSuffix(p, app) {
			// look at inside of this path if is directory (i hope)
			// search for deploy.json file
			// if found deploy json this is our project we looking for (i hope no conflict)
			_, err := os.Stat(path.Join(p, "deploy.json"))
			if err == nil {
				paths = append(paths, p)
			}

		}

		return nil
	})

	return paths
}

type Command struct {
	Command string `json:"command"`
	Args []string `json:"args"`
}

type DeployJSON struct {
	Name string `json:"name"`
	Build struct{
		Args []string `json:"args"`
      	} `json:"build"`
	PreBuild []Command `json:"pre-build"`
	PostBuild []Command `json:"post-build"`
}

func retrieveDeploy(p string) (*DeployJSON, error) {
	var deploy DeployJSON
	content, err := ioutil.ReadFile(path.Join(p, "deploy.json"))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(content, &deploy); err != nil {
		return nil, err
	}

	return &deploy, nil
}


func run(command Command, name string) {
	//build command
	fmt.Println(fmt.Sprintf("Executing %s commmand", command.Command))
	cmd := exec.Command(command.Command)
	for _, arg := range command.Args {
		cmd.Args = append(cmd.Args, strings.Replace(arg, "{$name}", name, -1))
	}
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(-1)
	}
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "we got an error trying to execute", command.Command)
		os.Exit(-1)
	}
	cmd.Wait()
}

func main() {

	args := os.Args[1:]
	if 0 == len(args) {
		fmt.Println("you dont give a application name to build")
		os.Exit(-1)
	}
	// first name of args must be application name

	gp, ok := os.LookupEnv("GOPATH")
	if !ok || 0 == len(gp) {
		fmt.Println("GOPATH is not set")
		os.Exit(-1)
	}

	paths := strings.Split(gp, fmt.Sprintf("%c", os.PathListSeparator))
	founded := make(map[string][]string)
	for _, p := range paths {
		// just in case
		if runtime.GOROOT() != p {
			p := path.Join(p, "src")
			for _, v := range findApplicationInPath(p, args[0]) {
				founded[p] = append(founded[p], v)
			}
		}
	}

	if  1 < len(founded) {
		fmt.Println(fmt.Sprintf("we found a conflict, %s inside of multiple GOPATH, please resolve this issue", args[0]))
		os.Exit(-1)
	}

	for gopath, projects := range founded {
		if 1 < len(projects) {
			fmt.Println(fmt.Sprintf("we found a conflict, multiple %s in same GOPATH; resolving this issue you make sure to", args[0]))
			fmt.Println("give more distinct name or path file project")
			os.Exit(-1)
		}
		// now everything is ok, we are gona read this deploy file and execute scenario's
		project := projects[0]
		deploy, err := retrieveDeploy(project)
		if err != nil  {
			fmt.Println(fmt.Sprintf("We got an error while reading deploy.json; %s", err.Error()))
			os.Exit(-1)
		}
		start := time.Now()

		// now we just make a loop for before build commands
		for _, command := range deploy.PreBuild {
			run(command, deploy.Name)
		}

		//build command
		full_name := strings.Replace(project, fmt.Sprintf("%s%c",gopath, os.PathSeparator), "", -1)
		cmd := exec.Command("go","build", "-o", deploy.Name)

		if 0 < len(deploy.Build.Args) {
			cmd.Args = append(cmd.Args, deploy.Build.Args...)
		}
		cmd.Args = append(cmd.Args, full_name)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("we got an error trying to execute go build")
			fmt.Println(output)
			os.Exit(-1)
		}

		// now we just make a loop for after build commands
		for _, command := range deploy.PostBuild {
			run(command, deploy.Name)
		}

		fmt.Println(fmt.Sprintf("deploying %s take %s", args[0], time.Since(start)))
		os.Exit(0)
	}

	fmt.Println(fmt.Sprintf("%s not found in GOPATH envs, make sure you have deploy.json in your project root", args[0]))
	os.Exit(-1)
}
