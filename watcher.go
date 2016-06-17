// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package watcher provides a tool for automatically running a custom command or
// running a custom method when detected file's changing.
package watcher

import (
	"os"
	"path/filepath"
	"github.com/fsnotify/fsnotify"
	"time"
	"os/exec"
	"fmt"
	"gopkg.in/orivil/log.v0"
	"os/signal"
	"syscall"
	"sync"
)

type AutoCommand struct {
	dirs       []string

	exts       map[string]bool

	watcher    *fsnotify.Watcher

	runChecker struct {
				   sync.Mutex
				   running bool
			   }

	errHandle  func(e error)
}

// Param exts provides the file name extensions for file watcher.
// Param errHandle is used to handle incoming errors.
func NewAutoCommand(exts []string, errHandle func(e error)) *AutoCommand {
	es := make(map[string]bool, len(exts))
	for _, ext := range exts {
		es[ext] = true
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	return &AutoCommand{
		exts: es,
		watcher: watcher,
		errHandle: errHandle,
	}
}

// Watch provides the directories which need to listen
func (b *AutoCommand) Watch(dir...string) {
	b.dirs = dir
}

// RunFunc runs the h when detected file changing. This function will wait
// for interrupt or terminate signal.
func (b *AutoCommand) RunFunc(h func()) {

	b.runChecker.Lock()
	defer b.runChecker.Unlock()
	if b.runChecker.running {
		b.errHandle(fmt.Errorf("wathcer.AutoCommand.RunFunc: the watcher is already running"))
		return
	} else {
		b.runChecker.running = true
	}

	c := b.listen()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {

		for _ = range c {

			h()
		}
	}()
	<-sig
	b.watcher.Close()
}

// RunCommand depends on RunFunc, the given command will be run when file
// watcher detected file changing.
func (b *AutoCommand) RunCommand(command string, args... string) {

	com := command
	for _, arg := range args {
		com += " " + arg
	}
	b.RunFunc(func() {

		log.Printf("[start] running command: [%s]", com)
		cmd := exec.Command(command, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Printf("[faild] got error: %v", err)
		} else {
			log.Printf("[ok]")
		}
	})
}

func (b *AutoCommand) listen() <-chan time.Time {
	for _, dir := range b.dirs {
		ds := b.walk(dir)
		for _, dir = range ds {
			b.watcher.Add(dir)
		}
	}

	timer := time.NewTimer(0)
	go func() {
		for {
			select {
			case evt := <-b.watcher.Events:
				fExt := filepath.Ext(evt.Name)
				if b.exts[fExt] {
					// trigger last event after 3 second
					timer.Reset(3 * time.Second)
				} else if fExt == "" {
					info, err := os.Stat(evt.Name)
					if err != nil {
						b.errHandle(fmt.Errorf("watcher.AutoCommand.listen(): %v", err))
					} else {
						if info.IsDir() {
							if evt.Op & fsnotify.Create == fsnotify.Create {
								dirs := b.walk(evt.Name)
								for _, dir := range dirs {
									b.watcher.Add(dir)
								}
							} else if evt.Op & fsnotify.Remove == fsnotify.Remove {
								dirs := b.walk(evt.Name)
								for _, dir := range dirs {
									b.watcher.Remove(dir)
								}
							}
						}
					}
				}
			case err := <-b.watcher.Errors:
				b.errHandle(fmt.Errorf("watcher.AutoCommand.listen(): %v", err))
			}
		}
	}()
	return timer.C
}

// walk returns all sub directories and current directory.
func (b *AutoCommand) walk(dir string) (dirs []string) {
	dirs = []string{dir}
	var walk filepath.WalkFunc = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	}

	err := filepath.Walk(dir, walk)
	if err != nil {
		b.errHandle(fmt.Errorf("watcher.AutoCommand.walk(): %v", err))
	}
	return
}