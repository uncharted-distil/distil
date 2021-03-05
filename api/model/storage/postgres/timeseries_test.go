//
//    Copyright © 2021 Uncharted Software Inc.
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	api "github.com/uncharted-distil/distil/api/model"
)

func TestRemoveDuplicates(t *testing.T) {
	data := []*api.TimeseriesObservation{
		{Time: 1, Value: 10},
		{Time: 2, Value: 20},
		{Time: 2, Value: 30},
		{Time: 3, Value: 40},
		{Time: 4, Value: 50},
		{Time: 4, Value: 60},
		{Time: 4, Value: 70},
		{Time: 5, Value: 80},
	}

	expected := []*api.TimeseriesObservation{
		{Time: 1, Value: 10},
		{Time: 2, Value: 50},
		{Time: 3, Value: 40},
		{Time: 4, Value: 180},
		{Time: 5, Value: 80},
	}

	result := removeDuplicates(data)
	assert.Equal(t, expected, result)
}

func TestRemoveDuplicatesNoDuplicates(t *testing.T) {

	data := []*api.TimeseriesObservation{
		{Time: 1, Value: 10},
		{Time: 2, Value: 20},
		{Time: 3, Value: 30},
		{Time: 4, Value: 40},
	}
	result := removeDuplicates(data)
	assert.Equal(t, data, result)
}

func TestRemoveDuplicatesAllDuplicates(t *testing.T) {

	data := []*api.TimeseriesObservation{
		{Time: 1, Value: 10},
		{Time: 1, Value: 20},
		{Time: 1, Value: 30},
		{Time: 1, Value: 40},
	}

	expected := []*api.TimeseriesObservation{
		{Time: 1, Value: 100},
	}

	result := removeDuplicates(data)
	assert.Equal(t, expected, result)
}
