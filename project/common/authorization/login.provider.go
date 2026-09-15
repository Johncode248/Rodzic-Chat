package authorizationpg

// LoginProviderArgument backs the mobile-only "log in with an
// already-verified email" flow (Google on Android, Apple on iOS) -- see
// project/service/transport/authorization/login.google.mobile.go and
// login.apple.mobile.go. Unlike LoginArgument there's no Password: the
// provider token itself (verified in the transport handler, before this
// argument is even built) is the proof of identity, so this only ever
// matches an *existing* Account by email -- it never creates one. A parent
// still has to go through a director's registration link at least once.
type LoginProviderArgument struct {
	Email      string
	DeviceId   string
	DeviceName string
	UserAgent  string
	IpAddress  string
}

// GoogleLoginRequest is the body of POST /public/login/google (Android
// only). IdToken is Android's native Google Sign-In ID token, requested
// with the platform's web client ID as "server client ID" -- see
// project/service/internal/oauthverify/google.go.
type GoogleLoginRequest struct {
	IdToken    string `json:"id_token,omitempty"`
	DeviceId   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
}

// AppleLoginRequest is the body of POST /public/login/apple (iOS only).
// IdentityToken is what AuthenticationServices hands back after a native
// Sign in with Apple flow -- see
// project/service/internal/oauthverify/apple.go.
type AppleLoginRequest struct {
	IdentityToken string `json:"identity_token,omitempty"`
	DeviceId      string `json:"device_id,omitempty"`
	DeviceName    string `json:"device_name,omitempty"`
}
