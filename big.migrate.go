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

}
