package nfz

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	nfzpg "github.com/Johncode248/Rodzic-Chat/project/common/nfz"
)

var ErrSavedInvalid = errors.New("queue_id and provider are required")

func ListSaved(ctx context.Context, accountId string) (nfzpg.SavedPlacesResponse, error) {
	rows, err := gormdb.CustomSelect[nfzpg.SavedPlace](ctx, "select * from nfz_saved_places where account_id = ? order by create_time desc limit 100", accountId)
	if err != nil {
		return nfzpg.SavedPlacesResponse{}, err
	}
	if rows == nil {
		rows = []nfzpg.SavedPlace{}
	}
	return nfzpg.SavedPlacesResponse{Places: rows}, nil
}

func clip(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

// AddSaved stores (or refreshes) a saved placówka.
func AddSaved(ctx context.Context, accountId string, p nfzpg.SavedPlace) error {
	if strings.TrimSpace(p.QueueId) == "" || strings.TrimSpace(p.Provider) == "" {
		return ErrSavedInvalid
	}
	return gormdb.GormDB.WithContext(ctx).Exec(`insert into nfz_saved_places
 (account_id, queue_id, specialty, benefit, provider, place, address, locality, phone, latitude, longitude, create_time)
 values (?,?,?,?,?,?,?,?,?,?,?,?)
 on conflict (account_id, queue_id) do update set specialty = excluded.specialty, benefit = excluded.benefit, provider = excluded.provider,
   place = excluded.place, address = excluded.address, locality = excluded.locality, phone = excluded.phone`,
		accountId, clip(p.QueueId, 100), clip(p.Specialty, 80), clip(p.Benefit, 200), clip(p.Provider, 200), clip(p.Place, 200), clip(p.Address, 200),
		clip(p.Locality, 100), clip(p.Phone, 40), p.Latitude, p.Longitude, time.Now()).Error
}

func RemoveSaved(ctx context.Context, accountId, queueId string) error {
	return gormdb.GormDB.WithContext(ctx).Exec("delete from nfz_saved_places where account_id = ? and queue_id = ?", accountId, queueId).Error
}
