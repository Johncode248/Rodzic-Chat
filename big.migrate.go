package main

import (
	"context"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	guardianclaimspg "github.com/Johncode248/Rodzic-Chat/project/common/guardianclaims"
	mealspg "github.com/Johncode248/Rodzic-Chat/project/common/meals"
	mediapg "github.com/Johncode248/Rodzic-Chat/project/common/media"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	passwordresetspg "github.com/Johncode248/Rodzic-Chat/project/common/passwordresets"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
	sessionspg "github.com/Johncode248/Rodzic-Chat/project/common/sessions"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
)

func BigMigrate() {

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

}
