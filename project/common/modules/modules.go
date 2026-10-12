// Package modules switches whole product modules on or off at build time.
//
// The mobile app currently serves personal (parent) accounts only. The placówka module -- organizations, groups,
// teachers, school chat/polls/sign-ups, payments, school events/absences/meals, media -- is kept in the code base but is
// not migrated (see big.migrate.go) and every query that would touch its tables is skipped while Organizations is
// false. To bring it back: uncomment the "PLACÓWKI" block in big.migrate.go and set Organizations = true.
package modules

// Organizations enables everything that needs the placówka tables.
const Organizations = false
