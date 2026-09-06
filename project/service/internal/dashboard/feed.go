package dashboard

import (
	"context"
	"fmt"
	"sort"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	dashboardpg "github.com/Johncode248/Rodzic-Chat/project/common/dashboard"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/absences"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
)

// Feed aggregates the director dashboard's activity feed from three
// already-existing, already-queryable sources -- recently joined parents,
// recent absence reports, and recent announcement-channel messages --
// merged and sorted by CreateTime desc, capped at limit. See
// dashboardpg.FeedItem's doc comment for why this reads at request time
// instead of maintaining a dedicated event log.
func Feed(ctx context.Context, organizationId string, limit int) (res dashboardpg.FeedResponse, err error) {
	perSourceLimit := limit
	if perSourceLimit <= 0 {
		perSourceLimit = 20
	}

	items := make([]dashboardpg.FeedItem, 0, perSourceLimit*3)

	newParents, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Limit:   perSourceLimit,
			OrderBy: databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending},
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "account_type", Operator: "=", Value: accountspg.AccountTypeParent.Ordinal()}}},
			},
		}.AddOrganizationAuth(organizationId),
	})
	if err != nil {
		return res, err
	}
	for _, account := range newParents.Accounts {
		items = append(items, dashboardpg.FeedItem{
			Type: "new_parent", Id: account.Id,
			Title:      fmt.Sprintf("Nowy rodzic: %s", account.Name),
			CreateTime: account.CreateTime,
		})
	}

	recentAbsences, err := absences.Select(ctx, absencespg.SelectArgument{
		Query: databasepg.Query{
			Limit:   perSourceLimit,
			OrderBy: databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending},
			Groups:  []databasepg.Group{{}},
		}.AddOrganizationAuth(organizationId),
	})
	if err != nil {
		return res, err
	}
	for _, absence := range recentAbsences.Absences {
		items = append(items, dashboardpg.FeedItem{
			Type: "absence", Id: absence.Id,
			Title:      "Zgłoszono nieobecność",
			CreateTime: absence.CreateTime,
		})
	}

	announcementChannels, err := chats.SelectChannels(ctx, chatspg.ChannelSelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "type", Operator: "=", Value: chatspg.ChannelTypeAnnouncement.Ordinal()}}},
			},
		}.AddOrganizationAuth(organizationId),
	})
	if err != nil {
		return res, err
	}
	if len(announcementChannels.Channels) > 0 {
		channelIds := make([]string, 0, len(announcementChannels.Channels))
		for _, channel := range announcementChannels.Channels {
			channelIds = append(channelIds, channel.Id)
		}

		recentAnnouncements, err := chats.SelectMessages(ctx, chatspg.MessageSelectArgument{
			Query: databasepg.Query{
				Limit:   perSourceLimit,
				OrderBy: databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending},
				Groups: []databasepg.Group{
					{Filters: []databasepg.Filter{{Key: "channel_id", Operator: "IN", Value: channelIds}}},
				},
			},
		})
		if err != nil {
			return res, err
		}
		for _, message := range recentAnnouncements.Messages {
			title := message.Content
			if len(title) > 100 {
				title = title[:100] + "…"
			}
			items = append(items, dashboardpg.FeedItem{
				Type: "announcement", Id: message.Id,
				Title:      title,
				CreateTime: message.CreateTime,
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreateTime.After(items[j].CreateTime)
	})

	if len(items) > perSourceLimit {
		items = items[:perSourceLimit]
	}

	return dashboardpg.FeedResponse{Items: items}, nil
}
