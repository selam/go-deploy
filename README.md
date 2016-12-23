# go-deploy

 go-deploy is simple deploy tool for deploying go apps. go-deploy has 3 main function,
 execute prebuild commands (if they exist in deploy.json), if all of them is successfully runs,
 then execute go build command, if build command success too then execute post build commands off course if they exists

# how it works
  when you execute go-deploy command from command line it searches from all your GOPATH's from your envrioment variable,
  you can have more than one directory in GOPATH and it's ok.
  when it found a project, if there is a file named deploy.json, it start parse and execute commands
  if go-deploy has a conflict like same project name under gitlab.com and github.com then go-deploy wants to more specific argument like "parentpath/project"

  go-deploy can search more deeply for multiple binaries in same project.

# syntax of deploy.json
deploy.json very simple file, if there is a {$name} variable on pre-build and post-build command args then it replace with name value of deploy.json
```json
{
    "name": "name of go app",
    "build": {
        "args": [""]
    },
    "pre-build": [
       {"command": "echo", "args": ["hello world"]}
    ],
    "post-build": [
       {"command": "rsync", "args": ["{$name}", "hostname:/path/to/"]},
       {"command": "ssh", "args": ["hostname", "supervisorctl restart {$name}"]}
    ]

}
```
