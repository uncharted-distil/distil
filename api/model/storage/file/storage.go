//
//   Copyright © 2021 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package file

import (
	"github.com/uncharted-distil/distil/api/model"
)

// Storage accesses the underlying datamart instance.
type Storage struct {
	folder string
}

// NewMetadataStorage returns a constructor for a metadata storage.
func NewMetadataStorage(folder string) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		return &Storage{
			folder: folder,
		}, nil
	}
}
