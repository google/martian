// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package martian

import (
	"fmt"
	log "github.com/golang/glog"
	"runtime"
	"strings"
)

// Infof logs an info message with caller information.
func Infof(format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	log.InfoDepth(1, fmt.Sprintf("%s: %s", funcName(), msg))
}

// Debugf logs a debug message with caller information.
func Debugf(format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	if log.V(2) {
		log.InfoDepth(1, fmt.Sprintf("%s: %s", funcName(), msg))
	}
}

// Errorf logs an error message with caller information.
func Errorf(format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	log.ErrorDepth(1, fmt.Sprintf("%s: %s", funcName(), msg))
}

// funcName finds a suitable function name for logging.
func funcName() string {
	pcs := make([]uintptr, 3)
	runtime.Callers(3, pcs)

	for _, pc := range pcs {
		f := runtime.FuncForPC(pc)
		if f == nil {
			continue
		}

		name := f.Name()
		return name[strings.LastIndex(name, "/")+1:]
	}

	return "anonymous"
}
