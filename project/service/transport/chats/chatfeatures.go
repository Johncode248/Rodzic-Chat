package chatsgt

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

const (
	consentNotificationBody = "prosi o Twoje potwierdzenie"
	signupNotificationBody  = "utworzył(a) listę zapisów"
)

// caller is the authenticated account behind a request, read once from the JWT context.
type caller struct {
	organizationId string
	accountId      string
	accountType    int
	isStaff        bool
}

func getCaller(ctx context.Context) (caller, error) {
	var c caller
	var err error
	if c.organizationId, err = credentials.GetOrganizationIdFromContext(ctx); err != nil || c.organizationId == "" {
		return c, errors.New("Error getting organizationId from context")
	}
	if c.accountId, err = credentials.GetUserIdFromContext(ctx); err != nil || c.accountId == "" {
		return c, errors.New("Error getting accountId from context")
	}
	if c.accountType, err = credentials.GetAccountTypeFromContext(ctx); err != nil {
		return c, errors.New("Error getting accountType from context")
	}
	c.isStaff = c.accountType == accountspg.AccountTypeDirector.Ordinal() ||
		c.accountType == accountspg.AccountTypeCoDirector.Ordinal() ||
		c.accountType == accountspg.AccountTypeTeacher.Ordinal()
	return c, nil
}

// postableChannel loads channelId and applies the checks every "create something in a channel"
// handler shares: the channel belongs to the caller's organization, and (for an Announcement
// channel) the Announcements feature is on. Who may actually post is enforced by chats.CreateMessage.
func postableChannel(ctx context.Context, c caller, channelId string) (chatspg.Channel, int, error) {
	result, err := chats.SelectChannels(ctx, chatspg.ChannelSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: channelId}}}},
		},
	})
	if err != nil {
		return chatspg.Channel{}, http.StatusInternalServerError, err
	}
	if len(result.Channels) == 0 || result.Channels[0].OrganizationId != c.organizationId {
		return chatspg.Channel{}, http.StatusForbidden, chats.ErrChannelNotInOrganization
	}
	if result.Channels[0].Type == chatspg.ChannelTypeAnnouncement {
		enabled, err := features.IsEnabled(ctx, c.organizationId, featurespg.FeatureAnnouncements)
		if err != nil {
			return chatspg.Channel{}, http.StatusInternalServerError, err
		}
		if !enabled {
			return chatspg.Channel{}, http.StatusForbidden, errors.New("feature not enabled for this organization")
		}
	}
	return result.Channels[0], http.StatusOK, nil
}

// readableMessage verifies messageId is in the caller's organization and that the caller may read
// its channel, returning both.
func readableMessage(ctx context.Context, c caller, messageId string) (chatspg.Message, chatspg.Channel, int, error) {
	if messageId == "" {
		return chatspg.Message{}, chatspg.Channel{}, http.StatusBadRequest, errors.New("message_id is required")
	}
	ok, err := messageInOrganization(ctx, messageId, c.organizationId)
	if err != nil {
		return chatspg.Message{}, chatspg.Channel{}, http.StatusInternalServerError, err
	}
	if !ok {
		return chatspg.Message{}, chatspg.Channel{}, http.StatusForbidden, chats.ErrChannelNotInOrganization
	}
	message, err := chats.LoadMessage(ctx, messageId)
	if err != nil {
		return chatspg.Message{}, chatspg.Channel{}, http.StatusNotFound, err
	}
	channel, found, err := channelForMessage(ctx, messageId)
	if err != nil {
		return chatspg.Message{}, chatspg.Channel{}, http.StatusInternalServerError, err
	}
	if !found {
		return chatspg.Message{}, chatspg.Channel{}, http.StatusNotFound, errors.New("channel not found")
	}
	if err = chats.CheckReadPermission(ctx, channel, c.accountId, c.accountType); err != nil {
		return chatspg.Message{}, chatspg.Channel{}, http.StatusForbidden, err
	}
	return message, channel, http.StatusOK, nil
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(v)
}

func createStatus(err error) int {
	if err == chats.ErrRateLimited {
		return http.StatusTooManyRequests
	}
	return http.StatusForbidden
}

// ---- Consents ---------------------------------------------------------------------------------

// CreateConsent handles POST /private/chats/consents (staff only, see module.go).
func CreateConsent(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.ConsentCreateRequest
		res     chatspg.ConsentCreateResponse
		err     error
	)
	ctx := logger.Enter(r.Context(), "chatsgt.CreateConsent", r.Body)
	defer logger.Exit(ctx, "chatsgt.CreateConsent", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	c, err := getCaller(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	channel, status, err := postableChannel(ctx, c, request.ChannelId)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	res, err = chats.CreateConsent(ctx, request.ChannelId, c.organizationId, c.accountId, request.Content, c.accountType)
	if err != nil {
		http.Error(w, err.Error(), createStatus(err))
		return
	}
	writeJSON(w, res)
	go notifyNewMessage(channel, c.accountId, res.Message, consentNotificationBody+": "+request.Content)
}

// ConfirmConsent handles POST /private/chats/consents/confirm.
func ConfirmConsent(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.ConsentMessageRequest
		err     error
	)
	ctx := logger.Enter(r.Context(), "chatsgt.ConfirmConsent", r.Body)
	defer logger.Exit(ctx, "chatsgt.ConfirmConsent", &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	c, err := getCaller(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	message, channel, status, err := readableMessage(ctx, c, request.MessageId)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	if err = chats.ConfirmConsent(ctx, request.MessageId, c.accountId); err != nil {
		code := http.StatusInternalServerError
		if err == chats.ErrNotConsentMessage {
			code = http.StatusBadRequest
		}
		http.Error(w, err.Error(), code)
		return
	}
	res, err := chats.ConsentStatus(ctx, message, channel, c.accountId, c.isStaff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, res)
}

// ConsentStatus handles GET /private/chats/consents/status?message_id=.
func ConsentStatus(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.ConsentStatusResponse
		err error
	)
	ctx := logger.Enter(r.Context(), "chatsgt.ConsentStatus")
	defer logger.Exit(ctx, "chatsgt.ConsentStatus", &res, &err)

	c, err := getCaller(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	message, channel, status, err := readableMessage(ctx, c, r.URL.Query().Get("message_id"))
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	res, err = chats.ConsentStatus(ctx, message, channel, c.accountId, c.isStaff)
	if err != nil {
		code := http.StatusInternalServerError
		if err == chats.ErrNotConsentMessage {
			code = http.StatusBadRequest
		}
		http.Error(w, err.Error(), code)
		return
	}
	writeJSON(w, res)
}

// RemindConsent handles POST /private/chats/consents/remind (staff only).
func RemindConsent(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.ConsentMessageRequest
		res     chatspg.ConsentRemindResponse
		err     error
	)
	ctx := logger.Enter(r.Context(), "chatsgt.RemindConsent", r.Body)
	defer logger.Exit(ctx, "chatsgt.RemindConsent", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	c, err := getCaller(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	message, channel, status, err := readableMessage(ctx, c, request.MessageId)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	res.Reminded, err = chats.RemindPending(ctx, message, channel)
	if err != nil {
		code := http.StatusInternalServerError
		if err == chats.ErrNotConsentMessage {
			code = http.StatusBadRequest
		}
		http.Error(w, err.Error(), code)
		return
	}
	writeJSON(w, res)
}

// ---- Signup lists -----------------------------------------------------------------------------

// CreateSignup handles POST /private/chats/signups.
func CreateSignup(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.SignupCreateRequest
		res     chatspg.SignupCreateResponse
		err     error
	)
	ctx := logger.Enter(r.Context(), "chatsgt.CreateSignup", r.Body)
	defer logger.Exit(ctx, "chatsgt.CreateSignup", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	c, err := getCaller(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	channel, status, err := postableChannel(ctx, c, request.ChannelId)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	res, err = chats.CreateSignup(ctx, request.ChannelId, c.organizationId, c.accountId, request, c.accountType)
	if err != nil {
		http.Error(w, err.Error(), createStatus(err))
		return
	}
	writeJSON(w, res)
	go notifyNewMessage(channel, c.accountId, res.Message, signupNotificationBody+": "+request.Title)
}

// SelectSignup handles GET /private/chats/signups?message_id=.
func SelectSignup(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.SignupDetailResponse
		err error
	)
	ctx := logger.Enter(r.Context(), "chatsgt.SelectSignup")
	defer logger.Exit(ctx, "chatsgt.SelectSignup", &res, &err)

	c, err := getCaller(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if _, _, status, err := readableMessage(ctx, c, r.URL.Query().Get("message_id")); err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	res, err = chats.GetSignupDetail(ctx, r.URL.Query().Get("message_id"), c.accountId)
	if err != nil {
		code := http.StatusInternalServerError
		if err == chats.ErrNotSignupMessage {
			code = http.StatusBadRequest
		}
		http.Error(w, err.Error(), code)
		return
	}
	writeJSON(w, res)
}

// slotAccess resolves slotId to its message and verifies the caller may read that message's channel.
func slotAccess(ctx context.Context, c caller, slotId string) (string, int, error) {
	messageId, err := chats.SlotMessageId(ctx, slotId)
	if err != nil {
		if err == chats.ErrSlotNotFound {
			return "", http.StatusNotFound, err
		}
		return "", http.StatusInternalServerError, err
	}
	if _, _, status, err := readableMessage(ctx, c, messageId); err != nil {
		return "", status, err
	}
	return messageId, http.StatusOK, nil
}

// ClaimSlot handles POST /private/chats/signups/claim.
func ClaimSlot(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.SignupClaimRequest
		err     error
	)
	ctx := logger.Enter(r.Context(), "chatsgt.ClaimSlot", r.Body)
	defer logger.Exit(ctx, "chatsgt.ClaimSlot", &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	c, err := getCaller(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	messageId, status, err := slotAccess(ctx, c, request.SlotId)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	if err = chats.ClaimSlot(ctx, request.SlotId, c.accountId); err != nil {
		code := http.StatusInternalServerError
		if err == chats.ErrSlotFull {
			code = http.StatusConflict
		}
		http.Error(w, err.Error(), code)
		return
	}
	res, err := chats.GetSignupDetail(ctx, messageId, c.accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, res)
}

// UnclaimSlot handles DELETE /private/chats/signups/claim?slot_id=.
func UnclaimSlot(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx := logger.Enter(r.Context(), "chatsgt.UnclaimSlot")
	defer logger.Exit(ctx, "chatsgt.UnclaimSlot", &err)

	c, err := getCaller(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	slotId := r.URL.Query().Get("slot_id")
	messageId, status, err := slotAccess(ctx, c, slotId)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	if err = chats.UnclaimSlot(ctx, slotId, c.accountId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := chats.GetSignupDetail(ctx, messageId, c.accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, res)
}
