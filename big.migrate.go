package main

import (
	"context"
	"fmt"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	announcementspg "github.com/Johncode248/Rodzic-Chat/project/common/announcements"
	avatarspg "github.com/Johncode248/Rodzic-Chat/project/common/avatars"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
	childhealthpg "github.com/Johncode248/Rodzic-Chat/project/common/childhealth"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	devicetokenspg "github.com/Johncode248/Rodzic-Chat/project/common/devicetokens"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	groupteacherspg "github.com/Johncode248/Rodzic-Chat/project/common/groupteachers"
	guardianclaimspg "github.com/Johncode248/Rodzic-Chat/project/common/guardianclaims"
	helprequestspg "github.com/Johncode248/Rodzic-Chat/project/common/helprequests"
	mealspg "github.com/Johncode248/Rodzic-Chat/project/common/meals"
	mediapg "github.com/Johncode248/Rodzic-Chat/project/common/media"
	nfzpg "github.com/Johncode248/Rodzic-Chat/project/common/nfz"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	organizationstripeaccountspg "github.com/Johncode248/Rodzic-Chat/project/common/organizationstripeaccounts"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
	passwordresetspg "github.com/Johncode248/Rodzic-Chat/project/common/passwordresets"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
	placespg "github.com/Johncode248/Rodzic-Chat/project/common/places"
	platformsettingspg "github.com/Johncode248/Rodzic-Chat/project/common/platformsettings"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
	sessionspg "github.com/Johncode248/Rodzic-Chat/project/common/sessions"
	shoppingpg "github.com/Johncode248/Rodzic-Chat/project/common/shopping"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
	todospg "github.com/Johncode248/Rodzic-Chat/project/common/todos"
)

// BigMigrate creates the schema of the product as it runs TODAY: the mobile app for personal (parent) accounts.
//
//	identity     accounts, credentials, sessions, password_resets, device_tokens, avatars
//	family       children (+ blood_type / health_note of the health card), parent_children, parent_invites
//	calendar     child_events, child_event_children, child_event_parents
//	tasks        todos, todo_snoozes, todo_prefs
//	health       child_allergies, child_doctors, child_measurements, child_episodes, child_episode_entries,
//	             child_medications, child_doses
//	shopping     shopping_lists, shopping_items
//	community    help_requests, help_replies, help_reply_votes, help_photos, category_follows, announcements
//	nfz          nfz_saved_places
//
// Everything that belongs to placówki (organizations, groups, teachers, school chat/polls, payments, school events,
// absences, meals, media, plus the dormant caregivers/places features) lives in migratePlacowkiModule below. It is kept
// compiling but NOT called -- see the commented call. Enabling it later is two steps: uncomment the call and set
// modules.Organizations = true (project/common/modules).
func BigMigrate() {
	ctx := context.Background()

	// Serializes migrations across every process connected to this database -- without it two instances booting
	// concurrently during a rolling deploy can race AutoMigrate's check-then-act ADD COLUMN/CREATE INDEX logic.
	lockConn, err := gormdb.AcquireMigrationLock(ctx)
	if err != nil {
		panic(err)
	}
	defer gormdb.ReleaseMigrationLock(ctx, lockConn)

	migrate(ctx,
		// identity -- accounts first: nearly every other table points at it.
		accountspg.Account{},
		credentialspg.Credential{},
		sessionspg.Session{},
		passwordresetspg.PasswordReset{},
		devicetokenspg.DeviceToken{},
		avatarspg.Avatar{},

		// family
		childrenpg.Child{},
		parentchildpg.ParentChild{},
		parentinvitespg.ParentInvite{},

		// calendar
		childeventspg.ChildEvent{},
		childeventspg.ChildEventChild{},
		childeventspg.ChildEventParent{},

		// tasks
		todospg.Todo{},
		todospg.Snooze{},
		todospg.Prefs{},

		// child health card + illness diary
		childhealthpg.Allergy{},
		childhealthpg.Doctor{},
		childhealthpg.Measurement{},
		childhealthpg.Episode{},
		childhealthpg.Entry{},
		childhealthpg.Medication{},
		childhealthpg.Dose{},

		// shopping lists
		shoppingpg.List{},
		shoppingpg.Item{},

		// community: "Szukam / Polecam", platform announcements, NFZ
		helprequestspg.HelpRequest{},
		helprequestspg.HelpReply{},
		helprequestspg.HelpReplyVote{},
		helprequestspg.HelpPhoto{},
		helprequestspg.CategoryFollow{},
		announcementspg.Announcement{},
		nfzpg.SavedPlace{},
	)

	applyConstraints(ctx)

	// PLACÓWKI -- not needed by the personal-account app. Uncomment together with modules.Organizations = true.
	// migratePlacowkiModule(ctx)
}

func migrate(ctx context.Context, models ...interface{}) {
	for _, m := range models {
		if err := gormdb.Migrate(ctx, m); err != nil {
			panic(err)
		}
	}
}

// constraint is a database-level rule GORM's AutoMigrate cannot express (or that would need an import cycle between
// model packages). Added idempotently, so BigMigrate stays safe to run on every boot.
type constraint struct {
	table, name, definition string
}

var constraints = []constraint{
	// Personal data is erased with its owner: hard-deleting an account (cmd/reaper) or a child removes everything below.
	{"credentials", "fk_credentials_account", "FOREIGN KEY (user_id) REFERENCES accounts(id) ON DELETE CASCADE"},
	{"parent_invites", "fk_parent_invites_inviter", "FOREIGN KEY (inviter_account_id) REFERENCES accounts(id) ON DELETE CASCADE"},
	{"nfz_saved_places", "fk_nfz_saved_places_account", "FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE"},
	{"child_events", "fk_child_events_creator", "FOREIGN KEY (created_by_account_id) REFERENCES accounts(id) ON DELETE CASCADE"},

	{"child_allergies", "fk_child_allergies_child", "FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE"},
	{"child_doctors", "fk_child_doctors_child", "FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE"},
	{"child_measurements", "fk_child_measurements_child", "FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE"},
	{"child_episodes", "fk_child_episodes_child", "FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE"},
	{"child_episode_entries", "fk_child_entries_episode", "FOREIGN KEY (episode_id) REFERENCES child_episodes(id) ON DELETE CASCADE"},
	{"child_medications", "fk_child_medications_child", "FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE"},
	{"child_doses", "fk_child_doses_medication", "FOREIGN KEY (medication_id) REFERENCES child_medications(id) ON DELETE CASCADE"},

	{"shopping_lists", "fk_shopping_lists_account", "FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE"},
	{"shopping_items", "fk_shopping_items_list", "FOREIGN KEY (list_id) REFERENCES shopping_lists(id) ON DELETE CASCADE"},

	// Value rules: the code validates too, the database is the last line of defence.
	{"children", "ck_children_blood_type", "CHECK (blood_type IS NULL OR blood_type IN ('', '0+', '0-', 'A+', 'A-', 'B+', 'B-', 'AB+', 'AB-'))"},
	{"child_allergies", "ck_child_allergies_severity", "CHECK (severity IN ('mild', 'moderate', 'severe'))"},
	{"child_episode_entries", "ck_child_entries_kind", "CHECK (kind IN ('temperature', 'symptom', 'note', 'medication'))"},
	{"child_medications", "ck_child_medications_interval", "CHECK (interval_hours BETWEEN 0 AND 48 AND min_gap_hours BETWEEN 0 AND 48)"},
	{"child_doses", "ck_child_doses_status", "CHECK (status IN ('pending', 'given', 'skipped'))"},
	{"shopping_lists", "ck_shopping_lists_status", "CHECK (status IN ('active', 'completed'))"},
}

func applyConstraints(ctx context.Context) {
	for _, c := range constraints {
		statement := fmt.Sprintf(`DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = '%s') THEN
    ALTER TABLE %s ADD CONSTRAINT %s %s;
  END IF;
END $$`, c.name, c.table, c.name, c.definition)
		if err := gormdb.GormDB.WithContext(ctx).Exec(statement).Error; err != nil {
			panic(fmt.Errorf("constraint %s on %s: %w", c.name, c.table, err))
		}
	}
}

// migratePlacowkiModule creates the placówka tables (organizations, memberships, groups, teachers, school chat and
// polls, payments, school events, absences, meals, media, invites) and the dormant parent-side extras that depend on
// them (caregivers, places, parent subscriptions). NOT called while the app serves personal accounts only.
func migratePlacowkiModule(ctx context.Context) {
	migrate(ctx,
		organizationspg.Organization{},
		accountorganizationspg.AccountOrganization{},
		featurespg.OrganizationFeature{},
		groupspg.Group{},
		groupteacherspg.GroupTeacher{},
		guardianclaimspg.GuardianClaim{},

		// school chat, polls, sign-ups
		chatspg.Channel{},
		chatspg.Message{},
		chatspg.Poll{},
		chatspg.PollOption{},
		chatspg.PollVote{},
		chatspg.MessageRead{},
		chatspg.MessageReaction{},
		chatspg.MessageConfirmation{},
		chatspg.SignupSlot{},
		chatspg.SignupClaim{},
		chatspg.ChannelMember{},
		chatspg.ChannelBan{},
		chatinvitespg.ChatInvite{},

		// school events, absences, meals, gallery
		eventspg.Event{},
		eventspg.EventRegistration{},
		absencespg.Absence{},
		mealspg.MealCancellation{},
		mediapg.Media{},

		// joining a placówka
		registrationlinkspg.RegistrationLink{},
		teacherinvitespg.TeacherInvite{},

		// payments
		platformsettingspg.PlatformSettings{},
		organizationstripeaccountspg.OrganizationStripeAccount{},
		paymentspg.Payment{},

		// parent-side features hidden in the current app
		parentsubscriptionspg.ParentSubscription{},
		caregiverspg.Caregiver{},
		caregiverspg.CaregiverRecommendation{},
		placespg.Place{},
		placespg.Recommendation{},
		placespg.GooglePlacesUsage{},
	)

	// AutoMigrate never drops an existing constraint. Channel/Message used to carry a foreign key to organizations;
	// org-less parent-to-parent chats use OrganizationId "" (no such row), so the insert would be rejected (SQLSTATE
	// 23503) on a database created before that FK was removed. Idempotent: a no-op where the constraint is already gone.
	for _, statement := range []string{
		"ALTER TABLE channels DROP CONSTRAINT IF EXISTS fk_channels_organization",
		"ALTER TABLE messages DROP CONSTRAINT IF EXISTS fk_messages_organization",
	} {
		if err := gormdb.GormDB.WithContext(ctx).Exec(statement).Error; err != nil {
			panic(err)
		}
	}
}
