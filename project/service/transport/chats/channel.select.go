package chatsgt

import (
	"encoding/json"
	"net/http"
	"strconv"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// maxChannelListLimit bounds ?limit= the same way the message-listing endpoints do -- without
// it ?limit=999999999 forces the backend to load (and JSON-encode) an organization's entire
// channel list in one request.
const maxChannelListLimit = 200

// SelectChannels lists channels in the caller's organization, newest-first, paginated via
// optional ?limit=/?offset= (default: no limit, i.e. everything -- existing callers that don't
// paginate keep working unchanged).
func SelectChannels(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.ChannelSelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.SelectChannels")
	defer logger.Exit(ctx, "chatsgt.SelectChannels", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	query := databasepg.Query{
		Groups:  []databasepg.Group{{}},
		OrderBy: databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending},
	}
	if l, convErr := strconv.Atoi(r.URL.Query().Get("limit")); convErr == nil && l > 0 {
		if l > maxChannelListLimit {
			l = maxChannelListLimit
		}
		query.Limit = l
	}
	if o, convErr := strconv.Atoi(r.URL.Query().Get("offset")); convErr == nil && o > 0 {
		query.Offset = o
	}
	query = query.AddOrganizationAuth(organizationId)

	res, err = chats.SelectChannels(ctx, chatspg.ChannelSelectArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accountType, accountTypeErr := credentials.GetAccountTypeFromContext(ctx)
	isStaff := accountTypeErr == nil && (accountType == accountspg.AccountTypeDirector.Ordinal() ||
		accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal())

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err == nil && accountId != "" {
		if !isStaff {
			// A parent may only see organization-wide channels plus
			// Group/Parents channels tied to a group their own child is
			// actually in -- otherwise any parent could enumerate every
			// other class's channel id from this list alone.
			parentGroupIds, groupErr := chats.ParentGroupIds(ctx, accountId)
			if groupErr != nil {
				http.Error(w, groupErr.Error(), http.StatusInternalServerError)
				return
			}
			allowedGroups := make(map[string]bool, len(parentGroupIds))
			for _, id := range parentGroupIds {
				allowedGroups[id] = true
			}

			visible := make([]chatspg.Channel, 0, len(res.Channels))
			for _, channel := range res.Channels {
				if (channel.Type == chatspg.ChannelTypeGroup || channel.Type == chatspg.ChannelTypeParents) &&
					channel.GroupId != "" && !allowedGroups[channel.GroupId] {
					continue
				}
				visible = append(visible, channel)
			}
			res.Channels = visible
		}

		res.Channels, err = chats.AttachUnreadFlags(ctx, res.Channels, accountId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(res)
}
