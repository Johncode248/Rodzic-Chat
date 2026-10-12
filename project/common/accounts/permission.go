package accountspg

// Permission is a bitmask of extra capabilities a director can grant to a
// Teacher account beyond what their AccountType alone allows -- e.g. a
// teacher who isn't a co-director but should still be able to create/manage
// payments (tuition, meals, trips...). Director and CoDirector accounts
// implicitly have every permission everywhere they're checked, so their
// Permissions bits are never read.
//
// New permissions must always be appended as the next power of two, never
// renumbered or reused -- this is persisted directly as an int in the DB,
// so shifting an existing bit would silently change what every account that
// already has it granted is allowed to do.
type Permission int64

const (
	PermissionManagePayments Permission = 1 << iota
	PermissionInviteParents
	PermissionResolveGuardianClaims
)

// PermissionInfo describes one grantable staff permission for the in-app explanation screen.
type PermissionInfo struct {
	Bit         Permission `json:"bit"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
}

// PermissionCatalog is the single source of truth for what each grantable permission means.
var PermissionCatalog = []PermissionInfo{
	{PermissionManagePayments, "manage_payments", "Płatności", "Tworzenie, edytowanie i usuwanie płatności (czesne, wycieczki, obiady) oraz rozliczanie wpłat."},
	{PermissionInviteParents, "invite_parents", "Zapraszanie rodziców", "Tworzenie i usuwanie linków rejestracyjnych, którymi rodzice dołączają do placówki."},
	{PermissionResolveGuardianClaims, "resolve_guardian_claims", "Potwierdzanie opiekunów", "Zatwierdzanie lub odrzucanie zgłoszeń rodziców, którzy twierdzą, że są opiekunem danego dziecka."},
}

// Has reports whether bit is set in the receiver bitmask.
func (p Permission) Has(bit Permission) bool {
	return p&bit != 0
}

// Set returns the bitmask with bit added (grant) or removed (revoke).
func (p Permission) Set(bit Permission, grant bool) Permission {
	if grant {
		return p | bit
	}
	return p &^ bit
}
