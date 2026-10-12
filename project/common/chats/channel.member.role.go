package chatspg

// ChannelMemberRole must always be appended after the last existing value,
// never inserted -- stored as a plain int ordinal, see channel.type.go for
// the same convention.
type ChannelMemberRole int

const (
	// ChannelMemberRoleMember can read, post, react and leave; anything more only through
	// individually granted permissions (see ChannelPermission).
	ChannelMemberRoleMember ChannelMemberRole = iota
	// ChannelMemberRoleAdmin is granted to whoever created the channel via a chat invite (see
	// project/common/chatinvites). Has every permission and is the only one who can change other
	// members' role/permissions. Exactly one per channel and never assignable -- the per-account
	// "3 channels" cap is counted from this role.
	ChannelMemberRoleAdmin
	// ChannelMemberRoleModerator is a helper the admin appoints: can invite, remove and temporarily
	// ban plain members, pin and delete messages -- but cannot rename the chat or manage roles.
	ChannelMemberRoleModerator
)

func (r ChannelMemberRole) Ordinal() int {
	return int(r)
}

func ChannelMemberRoleFromOrdinal(v int) ChannelMemberRole {
	return ChannelMemberRole(v)
}

// ChannelPermission is a bitmask of extra capabilities on a ChannelMember, on top of its role's
// preset (see RolePermissions). Only append new bits -- stored as an int64.
type ChannelPermission int64

const (
	// PermInvite: create invite codes to add people to the chat.
	PermInvite ChannelPermission = 1 << iota
	// PermRemoveMembers: remove ("wyrzuć") plain members.
	PermRemoveMembers
	// PermBan: temporarily (or permanently) block plain members and lift blocks.
	PermBan
	// PermPin: pin and unpin messages.
	PermPin
	// PermDeleteMessages: delete other members' messages.
	PermDeleteMessages
	// PermRename: change the chat's name.
	PermRename
)

// AllChannelPermissions is every defined bit.
const AllChannelPermissions = PermInvite | PermRemoveMembers | PermBan | PermPin | PermDeleteMessages | PermRename

// RolePermissions is the preset a role carries; a member's effective permissions are this OR-ed
// with the bits individually granted to them.
func RolePermissions(r ChannelMemberRole) ChannelPermission {
	switch r {
	case ChannelMemberRoleAdmin:
		return AllChannelPermissions
	case ChannelMemberRoleModerator:
		return PermInvite | PermRemoveMembers | PermBan | PermPin | PermDeleteMessages
	default:
		return 0
	}
}

func (p ChannelPermission) Has(bit ChannelPermission) bool { return p&bit == bit }

// Describes a role or permission for the clients' "what can this role do?" screens -- one source of
// truth (served by GET /private/chats/roles) so Android and iOS never drift apart.
type ChannelPermissionInfo struct {
	Bit         ChannelPermission `json:"bit"`
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
}

type ChannelRoleInfo struct {
	Role        ChannelMemberRole `json:"role"`
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	// Assignable roles can be given by the admin; Admin itself cannot.
	Assignable  bool              `json:"assignable"`
	Permissions ChannelPermission `json:"permissions"`
}

type ChannelRolesResponse struct {
	Roles       []ChannelRoleInfo       `json:"roles"`
	Permissions []ChannelPermissionInfo `json:"permissions"`
}

// ChannelRolesCatalog is the static catalog of roles and permissions with their Polish descriptions.
func ChannelRolesCatalog() ChannelRolesResponse {
	return ChannelRolesResponse{
		Roles: []ChannelRoleInfo{
			{Role: ChannelMemberRoleAdmin, Key: "admin", Name: "Administrator",
				Description: "Założyciel czatu. Może wszystko: zapraszać, usuwać i blokować osoby, przypinać i kasować wiadomości, zmieniać nazwę oraz nadawać role i uprawnienia innym. Jest jeden na czat i nie da się tej roli przekazać.",
				Permissions: RolePermissions(ChannelMemberRoleAdmin)},
			{Role: ChannelMemberRoleModerator, Key: "moderator", Name: "Moderator", Assignable: true,
				Description: "Pomocnik administratora. Może zapraszać osoby, usuwać i czasowo blokować zwykłych członków, przypinać i kasować wiadomości. Nie zmienia nazwy czatu ani ról.",
				Permissions: RolePermissions(ChannelMemberRoleModerator)},
			{Role: ChannelMemberRoleMember, Key: "member", Name: "Członek", Assignable: true,
				Description: "Czyta i pisze wiadomości, reaguje i może opuścić czat. Dodatkowe uprawnienia (np. przypinanie) administrator może nadać osobno.",
				Permissions: RolePermissions(ChannelMemberRoleMember)},
		},
		Permissions: []ChannelPermissionInfo{
			{Bit: PermInvite, Key: "invite", Name: "Zapraszanie", Description: "Tworzy kody zaproszeń, aby dodać nowe osoby do czatu."},
			{Bit: PermRemoveMembers, Key: "remove", Name: "Usuwanie członków", Description: "Usuwa zwykłych członków z czatu. Usunięta osoba może zostać zaproszona ponownie."},
			{Bit: PermBan, Key: "ban", Name: "Blokowanie", Description: "Blokuje zwykłych członków na wybrany czas (lub bezterminowo): nie widzą wiadomości ani nie mogą pisać, dopóki blokada trwa lub nie zostanie zdjęta."},
			{Bit: PermPin, Key: "pin", Name: "Przypinanie", Description: "Przypina ważne wiadomości na górze czatu i je odpina."},
			{Bit: PermDeleteMessages, Key: "delete", Name: "Kasowanie wiadomości", Description: "Kasuje wiadomości innych osób (własne może skasować każdy)."},
			{Bit: PermRename, Key: "rename", Name: "Zmiana nazwy", Description: "Zmienia nazwę czatu."},
		},
	}
}
