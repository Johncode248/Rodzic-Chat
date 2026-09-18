package transport

import (
	"log"
	"net/http"
	"time"

	"github.com/Johncode248/Rodzic-Chat/api_bucket"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal"
	absencesgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/absences"
	accountsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/accounts"
	aichatgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/aichat"
	authgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/authorization"
	calendargt "github.com/Johncode248/Rodzic-Chat/project/service/transport/calendar"
	caregiversgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/caregivers"
	chatinvitesgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/chatinvites"
	chatsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/chats"
	childeventsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/childevents"
	childrengt "github.com/Johncode248/Rodzic-Chat/project/service/transport/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	dashboardgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/dashboard"
	devicetokensgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/devicetokens"
	eventsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/events"
	featuresgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/features"
	groupsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/groups"
	groupteachersgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/groupteachers"
	guardianclaimsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/guardianclaims"
	mealsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/meals"
	mediagt "github.com/Johncode248/Rodzic-Chat/project/service/transport/media"
	nfzgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/nfz"
	organizationsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/organizations"
	organizationstripeaccountsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/organizationstripeaccounts"
	parentchildgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/parentchild"
	parentinvitesgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/parentinvites"
	parentsubscriptionsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/parentsubscriptions"
	paymentsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/payments"
	placesgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/places"
	platformsettingsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/platformsettings"
	registrationlinksgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/registrationlinks"
	stripewebhookgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/stripewebhook"
	teacherinvitesgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/variables"
	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go/v81"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

func initTracer() {
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
}

func Start() {
	api_bucket.Init()
	internal.Init()
	initTracer()

	// variables.STRIPESecret (like JWTSecret/SMTP*) is a mustEnv() --
	// missing STRIPE_SECRET_KEY now fails fast at boot rather than letting
	// every Stripe call fail one at a time later.
	stripe.Key = variables.STRIPESecret

	router := mux.NewRouter()

	public := router.PathPrefix("/public").Subrouter()
	private := router.PathPrefix("/private").Subrouter()

	// Authentication-sensitive endpoints previously had no request cap at
	// all -- unlimited attempts to brute-force a password, enumerate
	// registered emails via /password-recovery timing/response, or spam
	// the SMTP relay via /register. Each group gets its own limiter
	// (login attempts vs. registration vs. password-reset vs.
	// refresh-token, which legitimate clients call far more often) rather
	// than one shared budget across unrelated endpoints.
	loginRateLimit := credentials.RateLimitMiddleware(20, 5*time.Minute)
	registerRateLimit := credentials.RateLimitMiddleware(10, time.Hour)
	passwordResetRateLimit := credentials.RateLimitMiddleware(10, 15*time.Minute)
	refreshRateLimit := credentials.RateLimitMiddleware(60, 5*time.Minute)

	public.Handle("/register", registerRateLimit(http.HandlerFunc(authgt.Register))).Methods("POST")
	// A parent signing up entirely on their own, no director invite
	// involved (mobile-only client-side; see
	// authorizationpg.RegisterParentArgument's doc comment).
	public.Handle("/register/parent", registerRateLimit(http.HandlerFunc(authgt.RegisterParent))).Methods("POST")
	public.Handle("/login", loginRateLimit(http.HandlerFunc(authgt.Login))).Methods("POST")
	// Mobile-only alternate logins: Android sends a native Google Sign-In
	// ID token, iOS sends a native Sign in with Apple identity token. Both
	// verify the token server-side (project/service/internal/oauthverify)
	// and match an *existing* Account by its verified email -- distinct
	// from the /auth/google/login + /login/callback pair below, which is a
	// browser redirect flow for the web app, not for native mobile.
	public.Handle("/login/google", loginRateLimit(http.HandlerFunc(authgt.LoginGoogleMobile))).Methods("POST")
	public.Handle("/login/apple", loginRateLimit(http.HandlerFunc(authgt.LoginAppleMobile))).Methods("POST")
	public.Handle("/refresh-token", refreshRateLimit(http.HandlerFunc(authgt.RefreshToken))).Methods("POST")
	public.Handle("/password-recovery", passwordResetRateLimit(http.HandlerFunc(authgt.PasswordRecovery))).Methods("POST")
	public.Handle("/password-reset", passwordResetRateLimit(http.HandlerFunc(authgt.ConsumePasswordReset))).Methods("POST")

	public.HandleFunc("/auth/google/login", authgt.LoginGoogle).Methods("GET")
	public.HandleFunc("/auth/google/register", authgt.RegisterGoogle).Methods("GET")

	// Parent self-registration via a director-generated, group-scoped link
	// (see project/common/registrationlinks) — unauthenticated by design,
	// the token itself is the only "credential" needed to view/use it.
	public.HandleFunc("/registration-links/{token}", registrationlinksgt.PublicInfo).Methods("GET")
	public.HandleFunc("/registration-links/{token}/register", registrationlinksgt.PublicRegister).Methods("POST")

	// Teacher self-registration via a director-generated invite link (see
	// project/common/teacherinvites) — unauthenticated by design, same
	// reasoning as the parent registration links above.
	public.HandleFunc("/teacher-invites/{token}", teacherinvitesgt.PublicInfo).Methods("GET")
	public.HandleFunc("/teacher-invites/{token}/register", teacherinvitesgt.PublicRegister).Methods("POST")

	// One parent inviting a co-parent to manage the same children (see
	// project/common/parentinvites) — peer-to-peer, no organization/
	// director involved. /register is for someone with no account yet;
	// the authenticated /accept (below, under /private) is for someone
	// who already has one.
	public.HandleFunc("/parent-invites/{token}", parentinvitesgt.PublicInfo).Methods("GET")
	public.HandleFunc("/parent-invites/{token}/register", parentinvitesgt.PublicRegister).Methods("POST")

	router.HandleFunc("/login/callback", authgt.GoogleCallbackLoginHandler)
	router.HandleFunc("/register/callback", authgt.GoogleCallbackRegisterHandler)

	// Stripe calls this directly (no JWT, no session) -- authenticated by
	// verifying the Stripe-Signature header instead. Configure this URL
	// (<DOMEIN_B>/public/stripe/webhook) in the Stripe Dashboard once real
	// keys are in place.
	public.HandleFunc("/stripe/webhook", stripewebhookgt.Webhook).Methods("POST")

	// Platform-wide fee settings (project/common/platformsettings) -- no
	// org, no JWT; gated by its own shared-secret check (see
	// transport/platformsettings/helpers.go), so it's mounted under
	// /public rather than /private on purpose.
	public.HandleFunc("/platform/settings", platformsettingsgt.Select).Methods("GET")
	public.HandleFunc("/platform/settings", platformsettingsgt.Update).Methods("PATCH")

	// Exempting one organization from the platform subscription fee --
	// same shared-secret gate as platform/settings above (see
	// transport/organizations/subscription.override.go), never reachable
	// by a director's own JWT.
	public.HandleFunc("/organizations/subscription-override", organizationsgt.SetSubscriptionOverride).Methods("PATCH")

	private.Use(credentials.AuthMiddleware)

	private.HandleFunc("/credentials", credentials.Update).Methods("POST")

	// Logout revokes only the caller's own session (this device); it does
	// not affect any other device the account is logged in on.
	private.HandleFunc("/logout", authgt.Logout).Methods("POST")

	// "Who am I" + account self-service, scoped to the caller's own account id.
	private.HandleFunc("/accounts", accountsgt.Context).Methods("GET")
	private.HandleFunc("/accounts", accountsgt.Delete).Methods("DELETE")
	private.HandleFunc("/accounts", accountsgt.Update).Methods("PATCH")

	// Push notification device tokens: registered on login/app launch,
	// unregistered on logout. Any authenticated account, no feature gate --
	// this backs the chat push notifications, not a toggleable module.
	private.HandleFunc("/device-tokens", devicetokensgt.Register).Methods("POST")
	private.HandleFunc("/device-tokens", devicetokensgt.Unregister).Methods("DELETE")

	// Organizations: create is director-only (a director already gets an
	// organization via /register, this is exposed for completeness); read/update/
	// delete are always scoped to the caller's own JWT organizationId.
	// CoDirector has identical permissions to Director everywhere this
	// gate is used (per product decision) -- the only asymmetry is
	// provisioning: only a Director may invite a new CoDirector, checked
	// inside teacherinvitesgt.Create itself.
	organizationsDirectorOnly := credentials.RequireAccountType(accountspg.AccountTypeDirector.Ordinal(), accountspg.AccountTypeCoDirector.Ordinal())

	// Roster: staff-only "who is on my team" view, org-scoped, optional
	// ?account_type= filter, ?q= free-text search (name/login/email), and
	// ?status=removed to see soft-deleted accounts -- distinct from
	// /accounts' self-only "who am I".
	staffOnlyRoster := credentials.RequireAccountType(accountspg.AccountTypeDirector.Ordinal(), accountspg.AccountTypeCoDirector.Ordinal(), accountspg.AccountTypeTeacher.Ordinal())
	private.Handle("/accounts/roster", staffOnlyRoster(http.HandlerFunc(accountsgt.Roster))).Methods("GET")

	// Staff account deletion: director/co-director deletes a teacher or
	// co-director account within their own org (distinct from the
	// self-service DELETE /accounts above, which only ever targets the
	// caller's own account).
	private.Handle("/accounts/staff", organizationsDirectorOnly(http.HandlerFunc(accountsgt.DeleteStaff))).Methods("DELETE")

	// Staff permission grants: director/co-director grants or revokes extra
	// capabilities (currently: managing payments) on a teacher account in
	// their own org, without promoting them to co-director.
	private.Handle("/accounts/staff/permissions", organizationsDirectorOnly(http.HandlerFunc(accountsgt.UpdateStaffPermissions))).Methods("PATCH")

	// Director dashboard activity feed (see project/common/dashboard).
	private.Handle("/dashboard/feed", staffOnlyRoster(http.HandlerFunc(dashboardgt.Feed))).Methods("GET")

	// Guardian claims: review queue for the "second parent matched an
	// existing child by name but the birth date didn't agree" security
	// check (see project/common/guardianclaims) -- director-only.
	private.Handle("/guardian-claims", organizationsDirectorOnly(http.HandlerFunc(guardianclaimsgt.Select))).Methods("GET")
	private.Handle("/guardian-claims/resolve", organizationsDirectorOnly(http.HandlerFunc(guardianclaimsgt.Resolve))).Methods("POST")

	// Registration links: how a director lets parents in one group
	// self-register (see project/common/registrationlinks). Director-only
	// to create/list/revoke; the public GET/POST counterparts are
	// registered above under /public.
	private.Handle("/registration-links", organizationsDirectorOnly(http.HandlerFunc(registrationlinksgt.Create))).Methods("POST")
	private.Handle("/registration-links", organizationsDirectorOnly(http.HandlerFunc(registrationlinksgt.Select))).Methods("GET")
	private.Handle("/registration-links", organizationsDirectorOnly(http.HandlerFunc(registrationlinksgt.Delete))).Methods("DELETE")

	// Teacher invites: same shape as registration links but for a single
	// teacher (see project/common/teacherinvites) -- director-only to
	// create/list/revoke, public GET/POST counterparts registered above.
	private.Handle("/teacher-invites", organizationsDirectorOnly(http.HandlerFunc(teacherinvitesgt.Create))).Methods("POST")
	private.Handle("/teacher-invites", organizationsDirectorOnly(http.HandlerFunc(teacherinvitesgt.Select))).Methods("GET")
	private.Handle("/teacher-invites", organizationsDirectorOnly(http.HandlerFunc(teacherinvitesgt.Delete))).Methods("DELETE")

	private.Handle("/organizations", organizationsDirectorOnly(http.HandlerFunc(organizationsgt.Create))).Methods("POST")
	private.HandleFunc("/organizations", organizationsgt.Select).Methods("GET")
	private.Handle("/organizations", organizationsDirectorOnly(http.HandlerFunc(organizationsgt.Update))).Methods("PATCH")
	private.Handle("/organizations", organizationsDirectorOnly(http.HandlerFunc(organizationsgt.Delete))).Methods("DELETE")

	// Platform subscription billing: bank account + this organization's own
	// transfer-title reference + price computed from its ChildrenLimit tier
	// (see project/common/organizations/pricing.go). Director-only.
	private.Handle("/organizations/billing", organizationsDirectorOnly(http.HandlerFunc(organizationsgt.Billing))).Methods("GET")

	// Multi-facility account model: a parent can belong to several
	// organizations at once (see project/common/accountorganizations).
	private.HandleFunc("/organizations/mine", organizationsgt.Mine).Methods("GET")
	private.HandleFunc("/organizations/switch", authgt.SwitchOrganization).Methods("POST")
	private.Handle("/organizations/members", organizationsDirectorOnly(http.HandlerFunc(organizationsgt.RemoveMember))).Methods("DELETE")

	// Feature flags: director-only, scoped to the caller's own organization.
	private.Handle("/organizations/features", organizationsDirectorOnly(http.HandlerFunc(featuresgt.Create))).Methods("POST")
	// GET is open to any authenticated org member (not director-only): a
	// parent/teacher's app needs to know which features are on to render its
	// nav, and the handler already scopes results to the caller's own org.
	private.HandleFunc("/organizations/features", featuresgt.Select).Methods("GET")
	private.Handle("/organizations/features", organizationsDirectorOnly(http.HandlerFunc(featuresgt.Update))).Methods("PATCH")
	private.Handle("/organizations/features", organizationsDirectorOnly(http.HandlerFunc(featuresgt.Delete))).Methods("DELETE")

	// Children/Groups: only staff (director/co-director/teacher) can
	// create/update/delete within their own organization; any authenticated
	// org member can list.
	staffOnly := credentials.RequireAccountType(accountspg.AccountTypeDirector.Ordinal(), accountspg.AccountTypeCoDirector.Ordinal(), accountspg.AccountTypeTeacher.Ordinal())
	private.Handle("/children", staffOnly(http.HandlerFunc(childrengt.Create))).Methods("POST")
	private.HandleFunc("/children", childrengt.Select).Methods("GET")
	private.Handle("/children", staffOnly(http.HandlerFunc(childrengt.Update))).Methods("PATCH")
	private.Handle("/children", staffOnly(http.HandlerFunc(childrengt.Delete))).Methods("DELETE")

	// "Moje dzieci": a parent managing their own children directly, no
	// placowka required at all (see project/service/internal/children/mine.go)
	// -- distinct from the staff-only, org-scoped /children above. Any
	// authenticated account can use these (a parent adding their own kid),
	// scoped entirely to the caller via ParentChild, never to an org.
	private.HandleFunc("/children/mine", childrengt.CreateMine).Methods("POST")
	private.HandleFunc("/children/mine", childrengt.SelectMine).Methods("GET")
	private.HandleFunc("/children/mine", childrengt.UpdateMine).Methods("PATCH")
	private.HandleFunc("/children/mine/pesel", childrengt.RevealPesel).Methods("GET")
	private.HandleFunc("/children/mine", childrengt.DeleteMine).Methods("DELETE")

	// Personal, parent-created calendar entries per child (any linked
	// parent, no placowka involved) -- see project/common/childevents.
	private.HandleFunc("/child-events", childeventsgt.Create).Methods("POST")
	private.HandleFunc("/child-events", childeventsgt.Update).Methods("PATCH")
	private.HandleFunc("/child-events", childeventsgt.Delete).Methods("DELETE")

	// Merged calendar feed (personal child-events + relevant placowka
	// events) -- see project/service/internal/calendar.
	private.HandleFunc("/calendar", calendargt.Get).Methods("GET")

	// One parent inviting a co-parent to manage the same children -- the
	// authenticated "I already have an account" accept path; see
	// /public/parent-invites/{token}/register above for the no-account path.
	private.HandleFunc("/parent-invites", parentinvitesgt.Create).Methods("POST")
	private.HandleFunc("/parent-invites/{token}/accept", parentinvitesgt.Accept).Methods("POST")

	// Chat invites (project/common/chatinvites): a parent generates a
	// link/code for another parent to start a private 1:1 chat with no
	// shared placówka required -- the "rodzice między sobą mimo że szkoła
	// nie korzysta z aplikacji" case. Parent-only, and deliberately not
	// behind chatFeature below (there's no organization to check a feature
	// flag against).
	parentOnly := credentials.RequireAccountType(accountspg.AccountTypeParent.Ordinal())
	private.Handle("/chat-invites", parentOnly(http.HandlerFunc(chatinvitesgt.Create))).Methods("POST")
	private.Handle("/chat-invites/{token}/accept", parentOnly(http.HandlerFunc(chatinvitesgt.Accept))).Methods("POST")
	private.Handle("/chat-invites/for-channel", parentOnly(http.HandlerFunc(chatinvitesgt.CreateForChannel))).Methods("POST")

	// Parent AI plan (Free/Pro/Max) + usage -- see
	// project/common/parentsubscriptions. test-set-plan is TEMPORARY
	// scaffolding (see its own doc comment) standing in for real
	// StoreKit/Play Billing IAP, which isn't wired up yet.
	private.HandleFunc("/parent-subscription", parentsubscriptionsgt.Status).Methods("GET")
	private.HandleFunc("/parent-subscription/limits", parentsubscriptionsgt.Limits).Methods("GET")
	private.HandleFunc("/parent-subscription/test-set-plan", parentsubscriptionsgt.TestSetPlan).Methods("POST")

	// Parent AI assistant chat -- see project/service/internal/aichat.
	// Enforces the caller's plan/limit above, then calls Anthropic
	// (Claude). Mobile-only client-side.
	private.HandleFunc("/ai-chat/messages", aichatgt.Send).Methods("POST")

	// Parent-recommended places -- see project/common/places.
	private.HandleFunc("/places/categories", placesgt.Categories).Methods("GET")
	private.HandleFunc("/places/search", placesgt.Search).Methods("GET")
	private.HandleFunc("/places/recommend", placesgt.Recommend).Methods("POST")
	private.HandleFunc("/places/recommend", placesgt.Unrecommend).Methods("DELETE")
	private.HandleFunc("/places", placesgt.List).Methods("GET")

	// Nanny/tutor professional directory (project/common/caregivers) --
	// distinct from /places above: each row is owned by the account that
	// provides the service (self-listed), not an arbitrary third-party
	// place anyone can add.
	private.HandleFunc("/caregivers", caregiversgt.List).Methods("GET")
	private.HandleFunc("/caregivers", caregiversgt.Upsert).Methods("POST")
	private.HandleFunc("/caregivers", caregiversgt.Delete).Methods("DELETE")
	private.HandleFunc("/caregivers/mine", caregiversgt.GetMine).Methods("GET")
	private.HandleFunc("/caregivers/detail", caregiversgt.GetDetail).Methods("GET")
	private.HandleFunc("/caregivers/recommendations", caregiversgt.Recommend).Methods("POST")
	private.HandleFunc("/caregivers/recommendations", caregiversgt.Unrecommend).Methods("DELETE")

	// NFZ "Terminy leczenia" (wait-time forecasts) integration -- see project/common/nfz and
	// project/service/internal/nfz. Public NFZ API, no key needed server-side either; gated
	// behind our own auth like every other /private route regardless.
	private.HandleFunc("/nfz/benefits", nfzgt.FindBenefit).Methods("GET")
	private.HandleFunc("/nfz/queues", nfzgt.SearchQueues).Methods("GET")
	private.HandleFunc("/nfz/province", nfzgt.ProvinceForLocation).Methods("GET")

	// RSPO (Rejestr Szkół i Placówek Oświatowych) integration -- see project/common/rspo.
	// DELIBERATELY NOT ROUTED: rspo.Search (project/service/internal/rspo/client.go) calls
	// https://rspo.gov.pl/api/Institution, the undocumented internal endpoint the public
	// rspo.gov.pl search page's own JavaScript uses -- not the Ministry of Education's actual
	// sanctioned integration API (https://api.rspo.gov.pl), which requires a written "wniosek"
	// to rspo@cie.gov.pl and issued credentials before any programmatic use. Re-enable this
	// route (and the matching iOS hub section, see SearchHubView's doc comment) once that
	// request is approved and the client is pointed at the real credentialed endpoint.
	// private.HandleFunc("/rspo/search", rspogt.Search).Methods("GET") -- import commented out too, see above

	private.Handle("/groups", staffOnly(http.HandlerFunc(groupsgt.Create))).Methods("POST")
	private.Handle("/groups", staffOnly(http.HandlerFunc(groupsgt.Select))).Methods("GET")
	private.Handle("/groups", staffOnly(http.HandlerFunc(groupsgt.Update))).Methods("PATCH")
	private.Handle("/groups", staffOnly(http.HandlerFunc(groupsgt.Delete))).Methods("DELETE")

	// Payments: director/co-director create/edit/reconcile, or a teacher the
	// director has explicitly granted PermissionManagePayments (see
	// /accounts/staff/permissions above); any org member can list (a parent
	// only ever sees their own, see paymentsgt.Select). Checkout isn't
	// staff-gated -- it's implicitly scoped to "your own assigned payment"
	// inside payments.CreateCheckoutSession.
	paymentsFeature := credentials.RequireFeature(featurespg.FeaturePayments)
	paymentsManage := credentials.RequireAccountTypeOrPermission(
		accountspg.PermissionManagePayments,
		accountspg.AccountTypeDirector.Ordinal(), accountspg.AccountTypeCoDirector.Ordinal(),
	)
	private.Handle("/payments", paymentsFeature(paymentsManage(http.HandlerFunc(paymentsgt.Create)))).Methods("POST")
	private.Handle("/payments", paymentsFeature(http.HandlerFunc(paymentsgt.Select))).Methods("GET")
	private.Handle("/payments", paymentsFeature(paymentsManage(http.HandlerFunc(paymentsgt.Update)))).Methods("PATCH")
	private.Handle("/payments", paymentsFeature(paymentsManage(http.HandlerFunc(paymentsgt.Delete)))).Methods("DELETE")
	private.Handle("/payments/checkout", paymentsFeature(http.HandlerFunc(paymentsgt.Checkout))).Methods("POST")

	private.Handle("/organizations/stripe/onboarding-link", paymentsFeature(organizationsDirectorOnly(http.HandlerFunc(organizationstripeaccountsgt.OnboardingLink)))).Methods("POST")
	private.Handle("/organizations/stripe/status", paymentsFeature(http.HandlerFunc(organizationstripeaccountsgt.Status))).Methods("GET")

	private.Handle("/group-teachers", staffOnly(http.HandlerFunc(groupteachersgt.Create))).Methods("POST")
	private.HandleFunc("/group-teachers", groupteachersgt.Select).Methods("GET")
	private.Handle("/group-teachers", staffOnly(http.HandlerFunc(groupteachersgt.Delete))).Methods("DELETE")

	// Parent-child links: any authenticated org member may create/read/delete
	// (a parent registering their own child, or staff managing enrollments);
	// every write/read is scoped to the caller's organization via the child.
	private.HandleFunc("/parent-child", parentchildgt.Create).Methods("POST")
	private.HandleFunc("/parent-child", parentchildgt.Select).Methods("GET")
	private.HandleFunc("/parent-child", parentchildgt.Delete).Methods("DELETE")

	// Chat: channels are staff-managed; messages/polls/votes/reads/reactions
	// are open to any org member subject to the per-channel-type permission
	// checks and rate limiting enforced inside project/service/internal/chats.
	// Everything under /private/chats requires FeatureChat; message/poll
	// creation additionally checks FeatureAnnouncements at the handler level
	// when the target channel is an Announcement channel.
	chatFeature := credentials.RequireFeature(featurespg.FeatureChat)
	private.Handle("/chats/channels", chatFeature(staffOnly(http.HandlerFunc(chatsgt.CreateChannel)))).Methods("POST")
	private.Handle("/chats/channels", chatFeature(http.HandlerFunc(chatsgt.SelectChannels))).Methods("GET")
	private.Handle("/chats/channels", chatFeature(staffOnly(http.HandlerFunc(chatsgt.DeleteChannel)))).Methods("DELETE")

	private.Handle("/chats/messages", chatFeature(http.HandlerFunc(chatsgt.CreateMessage))).Methods("POST")
	private.Handle("/chats/messages", chatFeature(http.HandlerFunc(chatsgt.SelectMessages))).Methods("GET")
	private.Handle("/chats/messages", chatFeature(http.HandlerFunc(chatsgt.DeleteMessage))).Methods("DELETE")

	private.Handle("/chats/polls", chatFeature(http.HandlerFunc(chatsgt.CreatePoll))).Methods("POST")
	private.Handle("/chats/polls", chatFeature(http.HandlerFunc(chatsgt.SelectPoll))).Methods("GET")
	private.Handle("/chats/polls/votes", chatFeature(http.HandlerFunc(chatsgt.CastVote))).Methods("POST")
	private.Handle("/chats/polls/votes", chatFeature(http.HandlerFunc(chatsgt.RetractVote))).Methods("DELETE")

	private.Handle("/chats/messages/reads", chatFeature(http.HandlerFunc(chatsgt.MarkRead))).Methods("POST")
	private.Handle("/chats/messages/reads", chatFeature(http.HandlerFunc(chatsgt.SelectReads))).Methods("GET")

	private.Handle("/chats/messages/reactions", chatFeature(http.HandlerFunc(chatsgt.UpsertReaction))).Methods("POST")
	private.Handle("/chats/messages/reactions", chatFeature(http.HandlerFunc(chatsgt.SelectReactions))).Methods("GET")
	private.Handle("/chats/messages/reactions", chatFeature(http.HandlerFunc(chatsgt.DeleteReaction))).Methods("DELETE")

	// Org-less parent-to-parent chats (see chat-invites above and
	// parentchannel.*.go) -- authorized by chatspg.ChannelMember instead of
	// organization_id, so deliberately NOT wrapped in chatFeature: a parent
	// with no placówka at all still has these.
	private.HandleFunc("/chats/parent-channels", chatsgt.SelectParentChannels).Methods("GET")
	private.HandleFunc("/chats/parent-channels/messages", chatsgt.SelectParentChannelMessages).Methods("GET")
	private.HandleFunc("/chats/parent-channels/messages", chatsgt.CreateParentChannelMessage).Methods("POST")
	private.HandleFunc("/chats/parent-channels/members", chatsgt.SelectParentChannelMembers).Methods("GET")
	private.HandleFunc("/chats/parent-channels/members", chatsgt.RemoveParentChannelMember).Methods("DELETE")
	private.HandleFunc("/chats/parent-channels/rename", chatsgt.RenameParentChannel).Methods("PATCH")
	private.HandleFunc("/chats/parent-channels/messages/reads", chatsgt.MarkParentChannelMessageRead).Methods("POST")
	private.HandleFunc("/chats/parent-channels/messages/reads", chatsgt.SelectParentChannelMessageReads).Methods("GET")
	private.HandleFunc("/chats/parent-channels/messages/reactions", chatsgt.UpsertParentChannelMessageReaction).Methods("POST")
	private.HandleFunc("/chats/parent-channels/messages/reactions", chatsgt.SelectParentChannelMessageReactions).Methods("GET")
	private.HandleFunc("/chats/parent-channels/messages/reactions", chatsgt.DeleteParentChannelMessageReaction).Methods("DELETE")

	// Absences: any org member may report/list (ownership/membership is
	// checked in the handlers themselves -- a parent may only touch their
	// own children's absences), gated on FeatureAbsences.
	absencesFeature := credentials.RequireFeature(featurespg.FeatureAbsences)
	private.Handle("/absences", absencesFeature(http.HandlerFunc(absencesgt.Create))).Methods("POST")
	private.Handle("/absences", absencesFeature(http.HandlerFunc(absencesgt.Select))).Methods("GET")
	private.Handle("/absences", absencesFeature(http.HandlerFunc(absencesgt.Update))).Methods("PATCH")
	private.Handle("/absences", absencesFeature(http.HandlerFunc(absencesgt.Delete))).Methods("DELETE")

	// Events/trips: staff manage the events themselves; registrations
	// ("będzie / nie będzie") are declared by the child's parent or staff.
	// Gated on FeatureEvents.
	eventsFeature := credentials.RequireFeature(featurespg.FeatureEvents)
	private.Handle("/events", eventsFeature(staffOnly(http.HandlerFunc(eventsgt.CreateEvent)))).Methods("POST")
	private.Handle("/events", eventsFeature(http.HandlerFunc(eventsgt.SelectEvents))).Methods("GET")
	private.Handle("/events", eventsFeature(staffOnly(http.HandlerFunc(eventsgt.UpdateEvent)))).Methods("PATCH")
	private.Handle("/events", eventsFeature(staffOnly(http.HandlerFunc(eventsgt.DeleteEvent)))).Methods("DELETE")

	private.Handle("/events/registrations", eventsFeature(http.HandlerFunc(eventsgt.DeclareRegistration))).Methods("POST")
	private.Handle("/events/registrations", eventsFeature(http.HandlerFunc(eventsgt.SelectRegistrations))).Methods("GET")
	private.Handle("/events/registrations", eventsFeature(http.HandlerFunc(eventsgt.DeleteRegistration))).Methods("DELETE")

	// Gallery: uploader-or-staff can delete their own upload, only staff can
	// delete anyone's (checked in the handler). Gated on FeatureGallery.
	galleryFeature := credentials.RequireFeature(featurespg.FeatureGallery)
	private.Handle("/media", galleryFeature(http.HandlerFunc(mediagt.Upload))).Methods("POST")
	private.Handle("/media", galleryFeature(http.HandlerFunc(mediagt.Select))).Methods("GET")
	private.Handle("/media", galleryFeature(http.HandlerFunc(mediagt.Delete))).Methods("DELETE")

	// Meal cancellations ("odwołania obiadu"): same ownership shape as
	// absences -- any org member may report/list (a parent only ever
	// touches their own children's cancellations), gated on
	// FeatureMealCancellations (disabled by default per organization).
	mealsFeature := credentials.RequireFeature(featurespg.FeatureMealCancellations)
	private.Handle("/meals/cancellations", mealsFeature(http.HandlerFunc(mealsgt.Create))).Methods("POST")
	private.Handle("/meals/cancellations", mealsFeature(http.HandlerFunc(mealsgt.Select))).Methods("GET")
	private.Handle("/meals/cancellations", mealsFeature(http.HandlerFunc(mealsgt.Update))).Methods("PATCH")
	private.Handle("/meals/cancellations", mealsFeature(http.HandlerFunc(mealsgt.Delete))).Methods("DELETE")

	// CorsMiddleware must wrap MaxSizeMiddleware(router), not router
	// directly -- the previous version built `handler` from
	// MaxSizeMiddleware(router) and then immediately discarded it by
	// reassigning `handler = credentials.CorsMiddleware(router)` (wrapping
	// the bare router again instead of the already-wrapped handler), so the
	// 5 MB body limit was built but never actually reachable by any real
	// request -- every /private and /public endpoint accepted a request
	// body of unbounded size.
	handler := MaxSizeMiddleware(router)
	handler = credentials.CorsMiddleware(handler)

	log.Println("Starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func MaxSizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wyjątki dla uploadu plików: /private/files (legacy) i /private/media
		// (gallery photos), obie mogą przekraczać domyślny limit 5MB.
		if (r.URL.Path == "/private/files" || r.URL.Path == "/private/media") && r.Method == "POST" {
			next.ServeHTTP(w, r)
			return
		}

		// Dla całej reszty limit 5 MB
		// 5 * 1024 * 1024 bajtów
		r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)

		next.ServeHTTP(w, r)
	})
}
