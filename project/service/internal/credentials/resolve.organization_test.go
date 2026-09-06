package credentials

import (
	"reflect"
	"sort"
	"testing"
)

func TestPickActiveOrganization_NoMemberships_FallsBackToPrimary(t *testing.T) {
	active, all := pickActiveOrganization(nil, "org-1", "")

	if active != "org-1" {
		t.Fatalf("expected fallback to primaryOrganizationId, got %q", active)
	}
	if !reflect.DeepEqual(all, []string{"org-1"}) {
		t.Fatalf("expected OrganizationIds to be [org-1], got %v", all)
	}
}

func TestPickActiveOrganization_NoMemberships_NoPrimary(t *testing.T) {
	active, all := pickActiveOrganization(nil, "", "")

	if active != "" {
		t.Fatalf("expected empty active org, got %q", active)
	}
	if all != nil {
		t.Fatalf("expected nil OrganizationIds, got %v", all)
	}
}

func TestPickActiveOrganization_RequestedMember_Wins(t *testing.T) {
	memberships := []membership{
		{OrganizationId: "org-1", IsPrimary: true},
		{OrganizationId: "org-2", IsPrimary: false},
	}

	active, all := pickActiveOrganization(memberships, "org-1", "org-2")

	if active != "org-2" {
		t.Fatalf("expected requested org-2 to win, got %q", active)
	}
	sort.Strings(all)
	if !reflect.DeepEqual(all, []string{"org-1", "org-2"}) {
		t.Fatalf("expected all org ids, got %v", all)
	}
}

func TestPickActiveOrganization_RequestedNotAMember_FallsBackToPrimary(t *testing.T) {
	memberships := []membership{
		{OrganizationId: "org-1", IsPrimary: true},
		{OrganizationId: "org-2", IsPrimary: false},
	}

	// org-3 is not one of the account's memberships -- must not be honored.
	active, _ := pickActiveOrganization(memberships, "org-1", "org-3")

	if active != "org-1" {
		t.Fatalf("expected fallback to the primary membership, got %q", active)
	}
}

func TestPickActiveOrganization_NoRequest_UsesPrimaryMembership(t *testing.T) {
	memberships := []membership{
		{OrganizationId: "org-1", IsPrimary: false},
		{OrganizationId: "org-2", IsPrimary: true},
	}

	active, _ := pickActiveOrganization(memberships, "org-1", "")

	if active != "org-2" {
		t.Fatalf("expected the IsPrimary membership to win, got %q", active)
	}
}
