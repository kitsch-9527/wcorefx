package sec

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestFormatSIDAttributes(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		attr     uint32
		expected string
	}{
		{
			name:     "no flags",
			label:    "test",
			attr:     0,
			expected: "g:[       ]",
		},
		{
			name:     "mandatory",
			label:    "test",
			attr:     windows.SE_GROUP_MANDATORY,
			expected: "g:[M      ]",
		},
		{
			name:     "enabled by default + enabled",
			label:    "test",
			attr:     windows.SE_GROUP_ENABLED_BY_DEFAULT | windows.SE_GROUP_ENABLED,
			expected: "g:[ DE    ]",
		},
		{
			name:     "owner",
			label:    "test",
			attr:     windows.SE_GROUP_OWNER,
			expected: "g:[   O   ]",
		},
		{
			name:     "deny only",
			label:    "test",
			attr:     windows.SE_GROUP_USE_FOR_DENY_ONLY,
			expected: "g:[    U  ]",
		},
		{
			name:     "logon id",
			label:    "test",
			attr:     windows.SE_GROUP_LOGON_ID,
			expected: "g:[     L ]",
		},
		{
			name:     "resource",
			label:    "test",
			attr:     windows.SE_GROUP_RESOURCE,
			expected: "g:[      R]",
		},
		{
			name:     "all flags",
			label:    "test",
			attr:     windows.SE_GROUP_MANDATORY | windows.SE_GROUP_ENABLED_BY_DEFAULT | windows.SE_GROUP_ENABLED | windows.SE_GROUP_OWNER | windows.SE_GROUP_USE_FOR_DENY_ONLY | windows.SE_GROUP_LOGON_ID | windows.SE_GROUP_RESOURCE,
			expected: "g:[MDEOULR]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSIDAttributes(tt.label, tt.attr)
			if result != tt.expected {
				t.Errorf("FormatSIDAttributes(%q, 0x%x) = %q, want %q", tt.label, tt.attr, result, tt.expected)
			}
		})
	}
}

func TestFormatPrivilegeStatus(t *testing.T) {
	tests := []struct {
		name     string
		attr     uint32
		expected string
	}{
		{
			name:     "no flags",
			attr:     0,
			expected: "P:[    ]",
		},
		{
			name:     "enabled by default",
			attr:     windows.SE_PRIVILEGE_ENABLED_BY_DEFAULT,
			expected: "P:[D   ]",
		},
		{
			name:     "enabled",
			attr:     windows.SE_PRIVILEGE_ENABLED,
			expected: "P:[ E  ]",
		},
		{
			name:     "removed",
			attr:     windows.SE_PRIVILEGE_REMOVED,
			expected: "P:[  R ]",
		},
		{
			name:     "used for access",
			attr:     windows.SE_PRIVILEGE_USED_FOR_ACCESS,
			expected: "P:[   A]",
		},
		{
			name:     "enabled by default + enabled",
			attr:     windows.SE_PRIVILEGE_ENABLED_BY_DEFAULT | windows.SE_PRIVILEGE_ENABLED,
			expected: "P:[DE  ]",
		},
		{
			name:     "all flags",
			attr:     windows.SE_PRIVILEGE_ENABLED_BY_DEFAULT | windows.SE_PRIVILEGE_ENABLED | windows.SE_PRIVILEGE_REMOVED | windows.SE_PRIVILEGE_USED_FOR_ACCESS,
			expected: "P:[DERA]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPrivilegeStatus(tt.attr)
			if result != tt.expected {
				t.Errorf("FormatPrivilegeStatus(0x%x) = %q, want %q", tt.attr, result, tt.expected)
			}
		})
	}
}

func TestFormatJoinStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   NETSETUP_JOIN_STATUS
		expected string
	}{
		{name: "unknown", status: NetSetupUnknownStatus, expected: "NetSetupUnknownStatus"},
		{name: "unjoined", status: NetSetupUnjoined, expected: "NetSetupUnjoined"},
		{name: "workgroup", status: NetSetupWorkgroupName, expected: "NetSetupWorkgroupName"},
		{name: "domain", status: NetSetupDomainName, expected: "NetSetupDomainName"},
		{name: "unknown value", status: NETSETUP_JOIN_STATUS(99), expected: "NETSETUP_JOIN_STATUS(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatJoinStatus(tt.status)
			if result != tt.expected {
				t.Errorf("FormatJoinStatus(%d) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}
