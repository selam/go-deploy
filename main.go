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


type DeployJSON struct {
	Name string `json:"name"`
	Build struct{
		Args string `json:"args"`
	      } `json:"build"`
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

	if 0 == len(founded) {
		fmt.Printf("%s not found in GOPATH envs, make sure you have deploy.json in your project root\n", args[0])
		os.Exit(-1)
	}
	if  1 < len(founded) {
		fmt.Printf("we found a conflict, %s inside of multiple GOPATH, please resolve this issue\n", args[0])
		os.Exit(-1)
	}

	for gopath, projects := range founded {
		if 1 < len(projects) {
			fmt.Printf("we found a conflict, multiple %s in same GOPATH; resolving this issue you make sure to \ngive more distinct name or path file project\n", args[0])
			os.Exit(-1)
		}
		// now everything is ok, we are gona read this deploy file and execute scenario's
		project := projects[0]
		deploy, err := retrieveDeploy(project)
		if err != nil  {
			fmt.Printf("We got an error while reading deploy.json; %s\n", err.Error())
			os.Exit(-1)
		}
		// fully name of project
		//fmt.Println(fmt.Sprintf("%c", os.PathSeparator))
		fmt.Println(strings.Replace(project, fmt.Sprintf("%s%c",gopath, os.PathSeparator), "", -1))
		output, err := exec.Command("go", "build", "--o", deploy.Name, deploy.Build.Args, strings.Replace(project, fmt.Sprintf("%s%c",gopath, os.PathSeparator), "", -1)).CombinedOutput()

		if err != nil {
			fmt.Printf("we got an error trying to execute go build;\n\n\n%s", output)
			os.Exit(-1)
		}

		fmt.Printf("%v \n", deploy.Name)
	}

}
