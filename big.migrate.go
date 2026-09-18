package main

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	devicetokenspg "github.com/Johncode248/Rodzic-Chat/project/common/devicetokens"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	groupteacherspg "github.com/Johncode248/Rodzic-Chat/project/common/groupteachers"
	guardianclaimspg "github.com/Johncode248/Rodzic-Chat/project/common/guardianclaims"
	mealspg "github.com/Johncode248/Rodzic-Chat/project/common/meals"
	mediapg "github.com/Johncode248/Rodzic-Chat/project/common/media"
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
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
)

func BigMigrate() {
	ctx := context.Background()

	// Serializes migrations across every process connected to this
	// database (see AcquireMigrationLock's doc comment) -- without it, two
	// instances booting concurrently during a rolling deploy can race
	// AutoMigrate's own check-then-act ADD COLUMN/CREATE INDEX logic and
	// crash-loop on a Postgres "already exists" error.
	lockConn, err := gormdb.AcquireMigrationLock(ctx)
	if err != nil {
		panic(err)
	}
	defer gormdb.ReleaseMigrationLock(ctx, lockConn)

	if err := gormdb.Migrate(context.Background(), organizationspg.Organization{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), accountspg.Account{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), devicetokenspg.DeviceToken{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), accountorganizationspg.AccountOrganization{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), credentialspg.Credential{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), sessionspg.Session{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), passwordresetspg.PasswordReset{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), featurespg.OrganizationFeature{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), childrenpg.Child{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), groupspg.Group{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), groupteacherspg.GroupTeacher{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), parentchildpg.ParentChild{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), guardianclaimspg.GuardianClaim{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), chatspg.Channel{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), chatspg.Message{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), chatspg.Poll{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), chatspg.PollOption{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), chatspg.PollVote{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), chatspg.MessageRead{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), chatspg.MessageReaction{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), chatspg.ChannelMember{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), chatinvitespg.ChatInvite{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), caregiverspg.Caregiver{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), caregiverspg.CaregiverRecommendation{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), absencespg.Absence{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), eventspg.Event{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), eventspg.EventRegistration{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), mediapg.Media{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), mealspg.MealCancellation{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), registrationlinkspg.RegistrationLink{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), teacherinvitespg.TeacherInvite{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), platformsettingspg.PlatformSettings{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), organizationstripeaccountspg.OrganizationStripeAccount{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), paymentspg.Payment{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), parentsubscriptionspg.ParentSubscription{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), placespg.Place{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), placespg.Recommendation{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), placespg.GooglePlacesUsage{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), childeventspg.ChildEvent{}); err != nil {
		panic(err)
	}

	if err := gormdb.Migrate(context.Background(), parentinvitespg.ParentInvite{}); err != nil {
		panic(err)
	}

	//dropObsoleteColumns(ctx)
}

// dropObsoleteColumns removes columns whose Go struct fields were deleted outright (not just
// renamed) -- AutoMigrate only ever ADDs columns/indexes, it never drops one on its own, so a
// field removed from a struct would otherwise leave its column sitting in the DB forever. Each
// statement is IF EXISTS, so this is safe to run on every boot (a fresh DB or one that already
// had the column dropped both no-op cleanly) and safe to run alongside AutoMigrate under the
// same migration lock above.
func dropObsoleteColumns(ctx context.Context) {
	statements := []string{
		// Account.Login: replaced by Email as the sole login key (2026-09-17).
		`ALTER TABLE accounts DROP COLUMN IF EXISTS login`,
		// EventRegistration.ConsentGiven / Event.RequiresConsent: consent-gated event
		// registration isn't supported in MVP (2026-09-17).
		`ALTER TABLE event_registrations DROP COLUMN IF EXISTS consent_given`,
		`ALTER TABLE events DROP COLUMN IF EXISTS requires_consent`,
		// Group.AgeRange: unused, descriptive-only field (2026-09-17).
		`ALTER TABLE groups DROP COLUMN IF EXISTS age_range`,
		// ParentChild.CanPickup: written but never enforced anywhere (2026-09-17).
		`ALTER TABLE parent_children DROP COLUMN IF EXISTS can_pickup`,
		// ParentSubscription: stripped down to just Plan for MVP -- usage/quota tracking wasn't
		// enforcing anything real (AI assistant is hard-disabled server-side; Google Places
		// per-parent daily cap removed in favor of the shared global monthly budget alone)
		// (2026-09-17).
		`ALTER TABLE parent_subscriptions DROP COLUMN IF EXISTS ai_requests_used`,
		`ALTER TABLE parent_subscriptions DROP COLUMN IF EXISTS ai_tokens_used`,
		`ALTER TABLE parent_subscriptions DROP COLUMN IF EXISTS period_reset_at`,
		`ALTER TABLE parent_subscriptions DROP COLUMN IF EXISTS google_places_used_today`,
		`ALTER TABLE parent_subscriptions DROP COLUMN IF EXISTS google_places_usage_date`,
		`ALTER TABLE parent_subscriptions DROP COLUMN IF EXISTS stripe_customer_id`,
		`ALTER TABLE parent_subscriptions DROP COLUMN IF EXISTS stripe_subscription_id`,
		// Organization.Subdomain/LogoUrl/PrimaryColor: unused branding fields the backend never
		// consumed itself (2026-09-17).
		`ALTER TABLE organizations DROP COLUMN IF EXISTS subdomain`,
		`ALTER TABLE organizations DROP COLUMN IF EXISTS logo_url`,
		`ALTER TABLE organizations DROP COLUMN IF EXISTS primary_color`,
	}

	for _, statement := range statements {
		if err := gormdb.GormDB.WithContext(ctx).Exec(statement).Error; err != nil {
			panic(err)
		}
	}
}
