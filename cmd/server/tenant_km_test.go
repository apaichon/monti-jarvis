package main

import "testing"

func TestTenantKMAgentIDAllowed(t *testing.T) {
	tests := []struct {
		name           string
		agentID        string
		hasAssignments bool
		assignedIDs    []string
		want           bool
	}{
		{name: "assigned custom avatar", agentID: "SVOA", hasAssignments: true, assignedIDs: []string{"svoa"}, want: true},
		{name: "unassigned custom avatar", agentID: "svoa", hasAssignments: true, assignedIDs: []string{"ava"}, want: false},
		{name: "built-in demo fallback", agentID: "ava", hasAssignments: false, want: true},
		{name: "empty id", agentID: "", hasAssignments: false, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tenantKMAgentIDAllowed(tt.agentID, tt.hasAssignments, tt.assignedIDs); got != tt.want {
				t.Fatalf("tenantKMAgentIDAllowed(%q, %t, %v) = %t, want %t", tt.agentID, tt.hasAssignments, tt.assignedIDs, got, tt.want)
			}
		})
	}
}
