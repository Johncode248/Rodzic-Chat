package announcements

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	announcementspg "github.com/Johncode248/Rodzic-Chat/project/common/announcements"
)

// List returns every announcement, newest first -- visible to any authenticated account, not
// scoped to a placówka (see Announcement's doc comment: this is the developer's own channel to
// every parent using the app, not an org-specific one).
func List(ctx context.Context) (res announcementspg.ListResponse, err error) {
	rows, err := gormdb.Select[announcementspg.Announcement](ctx, databasepg.Query{
		OrderBy: databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending},
	})
	if err != nil {
		return res, err
	}
	return announcementspg.ListResponse{Announcements: rows}, nil
}
