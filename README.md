# Orivil File Watcher

## Introduction

A tool for automatically running a custom command or running a custom method when detected file's changing.

## Install

go get -v gopkg.in/orivil/watcher.v0


## GO Auto Builder Example:

autobuild.go:

```GO
package main

import (
	"gopkg.in/orivil/watcher.v0"
	"log"
	"os"
	"strings"
	"path/filepath"
)

func main() {
	// watch ".go" file
	extensions := []string{".go"}

	// handle incoming errors
	var errHandler = func(e error) {

		log.Println(e)
	}

	runner := watcher.NewAutoCommand(extensions, errHandler)

	// watch library directories
	goPath, _ := os.LookupEnv("GOPATH")
	goPaths := strings.Split(goPath, ";")
	for _, path := range goPaths {
		if path != "" {
			runner.Watch(filepath.Join(path, "src"))
		}
	}

	// build current directory
	buildFile := "."

	// run the watcher and wait for event.
	runner.RunCommand("go", "build", buildFile)
}
```

Open terminal: `go install autobuild.go`.

Then you can use command "autobuild" under your project directory to build your project automatically.

> **Note:** If command "autobuild" does not exist, add the path "$GOPATH/bin" to your PATH environment variable.


## Contributors

https://github.com/orivil/watcher/graphs/contributors

## License

Released under the [MIT License](https://github.com/orivil/watcher/blob/master/LICENSE).