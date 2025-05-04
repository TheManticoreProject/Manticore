package nt_status_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

func TestNTStatusString(t *testing.T) {
	tests := []struct {
		name     string
		status   nt_status.NT_STATUS
		expected string
	}{
		{
			name:     "NT_STATUS_SUCCESS",
			status:   nt_status.NT_STATUS_SUCCESS,
			expected: "SUCCESS",
		},
		{
			name:     "NT_STATUS_PENDING",
			status:   nt_status.NT_STATUS_PENDING,
			expected: "PENDING",
		},
		{
			name:     "NT_STATUS_TIMEOUT",
			status:   nt_status.NT_STATUS_TIMEOUT,
			expected: "TIMEOUT",
		},
		{
			name:     "NT_STATUS_BUFFER_ALL_ZEROS",
			status:   nt_status.NT_STATUS_BUFFER_ALL_ZEROS,
			expected: "BUFFER_ALL_ZEROS",
		},
		{
			name:     "Unknown status",
			status:   nt_status.NT_STATUS(0xFFFFFFFF),
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNTStatusError(t *testing.T) {
	tests := []struct {
		name          string
		status        nt_status.NT_STATUS
		expectedError bool
	}{
		{
			name:          "NT_STATUS_SUCCESS",
			status:        nt_status.NT_STATUS_SUCCESS,
			expectedError: false,
		},
		{
			name:          "NT_STATUS_PENDING",
			status:        nt_status.NT_STATUS_PENDING,
			expectedError: true,
		},
		{
			name:          "NT_STATUS_TIMEOUT",
			status:        nt_status.NT_STATUS_TIMEOUT,
			expectedError: true,
		},
		{
			name:          "Unknown status",
			status:        nt_status.NT_STATUS(0xFFFFFFFF),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.status.Error()
			if (err != nil) != tt.expectedError {
				t.Errorf("Error() returned %v, expectedError: %v", err, tt.expectedError)
			}
		})
	}
}
