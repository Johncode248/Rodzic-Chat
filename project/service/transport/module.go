package transport

import (
	"log"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/api_bucket"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal"
	absencesgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/absences"
	accountsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/accounts"
	authgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/authorization"
	chatsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/chats"
	childrengt "github.com/Johncode248/Rodzic-Chat/project/service/transport/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	dashboardgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/dashboard"
	eventsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/events"
	featuresgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/features"
	groupsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/groups"
	guardianclaimsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/guardianclaims"
	mealsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/meals"
	mediagt "github.com/Johncode248/Rodzic-Chat/project/service/transport/media"
	organizationsgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/organizations"
	parentchildgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/parentchild"
	registrationlinksgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/registrationlinks"
	teacherinvitesgt "github.com/Johncode248/Rodzic-Chat/project/service/transport/teacherinvites"
	"github.com/gorilla/mux"
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

	router := mux.NewRouter()

	public := router.PathPrefix("/public").Subrouter()
	private := router.PathPrefix("/private").Subrouter()

	public.HandleFunc("/register", authgt.Register).Methods("POST")
	public.HandleFunc("/login", authgt.Login).Methods("POST")
	public.HandleFunc("/refresh-token", authgt.RefreshToken).Methods("POST")
	public.HandleFunc("/password-recovery", authgt.PasswordRecovery).Methods("POST")
	public.HandleFunc("/password-reset", authgt.ConsumePasswordReset).Methods("POST")

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

	router.HandleFunc("/login/callback", authgt.GoogleCallbackLoginHandler)
	router.HandleFunc("/register/callback", authgt.GoogleCallbackRegisterHandler)

	private.Use(credentials.AuthMiddleware)

	private.HandleFunc("/credentials", credentials.Update).Methods("POST")

	// Logout revokes only the caller's own session (this device); it does
	// not affect any other device the account is logged in on.
	private.HandleFunc("/logout", authgt.Logout).Methods("POST")

	// "Who am I" + account self-service, scoped to the caller's own account id.
	private.HandleFunc("/accounts", accountsgt.Context).Methods("GET")
	private.HandleFunc("/accounts", accountsgt.Delete).Methods("DELETE")
	private.HandleFunc("/accounts", accountsgt.Update).Methods("PATCH")

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

	private.Handle("/groups", staffOnly(http.HandlerFunc(groupsgt.Create))).Methods("POST")
	private.HandleFunc("/groups", groupsgt.Select).Methods("GET")
	private.Handle("/groups", staffOnly(http.HandlerFunc(groupsgt.Update))).Methods("PATCH")
	private.Handle("/groups", staffOnly(http.HandlerFunc(groupsgt.Delete))).Methods("DELETE")

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

	handler := MaxSizeMiddleware(router) // Dodajemy nasz limit
	handler = credentials.CorsMiddleware(router)

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
