/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

/*
The index system is a filesystem based data storage that uses a directory tree
approach in which index 'keys' (typically a name or word) are converted into paths
in the directory tree. Each character of the index is a single directory in the path.

Each path typically contains a single data group (like one user for example).
Data files in the index system are directly accessible via their index paths and
the tree structure can be quickly searched by prefix using typical directory
walking functions.

*/

//
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func pathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	return true
}

func readDirNames(dirname string, limit int) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(limit)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

type WalkFunc func(path string, info os.FileInfo, err error) error

func walk(path string, limit int, info os.FileInfo, walkFn WalkFunc) error {
	err := walkFn(path, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}
	names, err := readDirNames(path, limit)
	if err != nil {
		return walkFn(path, info, err)
	}
	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := os.Lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walk(filename, limit, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

func Walk(root string, limit int, walkFn WalkFunc) error {
	info, err := os.Lstat(root)
	if err != nil {
		return walkFn(root, nil, err)
	}
	return walk(root, limit, info, walkFn)
}

func WalkBranch(path string, limit int) []string {
	var results []string
	visit := func(fpath string, fi os.FileInfo, err error) (e error) {
		if err != nil {
			return err
		}
		if fpath == path || fi.IsDir() {
			return nil
		}
		results = append(results, fpath)
		return
	}
	Walk(path, limit, visit)
	return results
}

func makePath(path string) error {
	return os.MkdirAll(path, 0700)
}

func indexPath(name []byte) string {
	name = []byte(strings.ToLower(string(name)))
	b := bytes.Split(name, []byte{})
	return string(bytes.Join(b, []byte(string(os.PathSeparator))))
}
