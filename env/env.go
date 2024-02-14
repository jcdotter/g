// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://github.com/jcdotter/go/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package env

import (
	"io"
	"os"

	"github.com/jcdotter/go/errors"
)

// -----------------------------------------------------------------------------
// ERRORS

var (
	ErrFileNotFound = errors.NotFound("environment file")
)

// -----------------------------------------------------------------------------
// STUBS
// funcs available in the os package

func Get(key string) string {
	return os.Getenv(key)
}

func Set(key, value string) error {
	return os.Setenv(key, value)
}

func Unset(key string) error {
	return os.Unsetenv(key)
}

func Clear() {
	os.Clearenv()
}

func List() []string {
	return os.Environ()
}

// -----------------------------------------------------------------------------
// ENV

// env represents the environment variables and provides methods for
// getting, setting, and unsetting variables, inlcuding methods for
// loading variables from files.
type env map[string]string

func New() *env {
	e := make(env)
	return &e
}

// Env returns a representation of the environment variables which
// provides methods for getting, setting, and unsetting variables,
// inlcuding methods for loading variables from files.
func Env() *env {
	e := New()
	// TODO: load the environment variables
	// need function to parse the environment variables from strings
	return e
}

func (e *env) Load(filenames ...string) (err error) {
	for _, file := range files(filenames...) {
		if err = e.load(file); err != nil {
			return
		}
	}
	return ErrFileNotFound
}

func (e *env) load(file string) (err error) {
	var f *os.File
	if f, err = os.Open(file); os.IsNotExist(err) {
		return
	}
	defer f.Close()
	var b []byte
	if b, err = io.ReadAll(f); err != nil {
		return
	}
	*e = parse(b)
	return
}

// Load loads the environment variables from the specified files
// and sets them in the current environment. If no files are
// specified, Load will attempt to load the .env file in the
// current directory.
func Load(filenames ...string) (err error) {
	return New().Load(filenames...)
}

// files returns the list of files to load
func files(filenames ...string) (files []string) {
	if len(filenames) == 0 {
		return []string{".env"}
	}
	return filenames
}
