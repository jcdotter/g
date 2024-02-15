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

package pub

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("preparing publication")
	// parse module
	var f *os.File
	var err error
	var b []byte
	if f, err = os.Open("go.mod"); err != nil {
		fmt.Println("go.mod not found. cannot publish module.")
		return
	}
	if b, err = io.ReadAll(f); err != nil {
		fmt.Println("error reading go.mod")
		return
	}
	for i := 0; i < len(b); i++ {
		if b[i] == '\n' {
			i++
			if len(b) < i+7 {
				fmt.Println("module not found in go.mod")
				return
			}
			if string(b[i:i+7]) == "module " {
				i += 7
				for pos := i; i < len(b); i++ {
					if c := b[i]; c == '\n' || c == ' ' || c == '\t' || c == '\r' || (c == '/' && b[i+1] == '/') {
						mod := string(b[pos:i])
						fmt.Println("module:", mod)
						break
					}
				}
			}
		}
	}
	// get version
	mod := os.Args[1]
	ver := os.Args[2]
	exec.Command("git add .")
	exec.Command(fmt.Sprintf("git commit -m 'v%s'", ver))
	exec.Command(fmt.Sprintf("git tag v%s", ver))
	exec.Command("git push origin v" + ver)
	exec.Command(fmt.Sprintf("GOPROXY=proxy.golang.org go list -m github.com/jcdotter/%s@v%s", mod, ver))
	_ = `
git tag v0.1.0
git push origin v0.1.0
GOPROXY=proxy.golang.org go list -m example.com/mymodule@v0.1.0
	`
}

func buildCmds() string {

}

func getMod() string {

}
