/*
Copyright 2024.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

(C) Copyright Hewlett Packard Enterprise Development LP
*/

package tapms

import (
	"testing"

	"github.hpe.com/hpe/sshot-net-operator/models"
)

func TestValidateVNIRequestData(t *testing.T) {
	tests := []struct {
		name          string
		input         models.VNIRequestData
		expectedError string
	}{
		{
			name: "Valid VNICount and VNIRange",
			input: models.VNIRequestData{
				VNICount: 10,
				VNIRange: []string{"100-200"},
			},
			expectedError: "",
		},
		{
			name: "Invalid VNICount (negative)",
			input: models.VNIRequestData{
				VNICount: -1,
			},
			expectedError: "VNI count is invalid: -1",
		},
		{
			name: "Invalid VNICount (exceeds limit)",
			input: models.VNIRequestData{
				VNICount: 65537,
			},
			expectedError: "VNI count is invalid: 65537",
		},
		{
			name: "Invalid VNIRange (out of range)",
			input: models.VNIRequestData{
				VNICount: 10,
				VNIRange: []string{"0-70000"},
			},
			expectedError: "VNI range is invalid: [0-70000]",
		},
		{
			name: "Valid empty VNIRange",
			input: models.VNIRequestData{
				VNICount: 10,
				VNIRange: []string{},
			},
			expectedError: "",
		},
		{
			name: "Invalid VNIRange (reversed range)",
			input: models.VNIRequestData{
				VNICount: 10,
				VNIRange: []string{"70000-100"},
			},
			expectedError: "VNI range is invalid: [70000-100]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVNIRequestData(tt.input)
			if err != nil && err.Error() != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if err == nil && tt.expectedError != "" {
				t.Errorf("expected error %v, got no error", tt.expectedError)
			}
		})
	}
}
