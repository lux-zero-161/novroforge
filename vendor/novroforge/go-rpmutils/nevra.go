/*
 * Copyright (c) SAS Institute, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rpmutils

import (
	"fmt"
)

// NEVRA holds the name, epoch, version, release and arch of the RPM, which uniquely identifies it
type NEVRA struct {
	Name    string
	Epoch   string
	Version string
	Release string
	Arch    string
}

// TODO: in v2 change epoch to an int

func (nevra *NEVRA) String() string {
	return fmt.Sprintf("%s-%s:%s-%s.%s.rpm", nevra.Name, nevra.Epoch, nevra.Version, nevra.Release, nevra.Arch)
}
