package helprequestspg

import (
	"time"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
)

// Categories is the single source of truth for the "Szukam / Polecam" board categories.
var Categories = []string{
	"Wózek i fotelik", "Rower i hulajnoga", "Zabawki i książki", "Ubrania i buty", "Wyposażenie domu i pokoju",
	"Lekarz i specjalista", "Niania", "Korepetytor", "Zajęcia dla dzieci", "Miejsca i atrakcje", "Inne",
}

// HelpRequest is a "Szukam" post: someone asks the community for a recommendation.
type HelpRequest struct {
	Id          string              `json:"id,omitempty" gorm:"primaryKey"`
	AccountId   string              `json:"account_id,omitempty" gorm:"not null;index"`
	Account     *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Title       string              `json:"title,omitempty" gorm:"not null"`
	Category    string              `json:"category,omitempty" gorm:"not null;index"`
	Description string              `json:"description,omitempty"`
	City        string              `json:"city,omitempty" gorm:"index"`
	Closed      bool                `json:"closed" gorm:"not null;default:false"`
	// PhotoId optionally points at a HelpPhoto attached to the question.
	PhotoId string `json:"photo_id,omitempty"`
	// BestReplyId is the recommendation the author marked "To mi pomogło"; Solved is true while it is set.
	// Solved questions stay in the searchable catalog.
	BestReplyId string `json:"best_reply_id,omitempty"`
	Solved      bool   `json:"solved" gorm:"not null;default:false;index"`
	// OwnerSeenTime is when the author last opened the question: replies newer than it are "new" for the bell/push.
	OwnerSeenTime time.Time `json:"-"`
	CreateTime    time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime    time.Time `json:"update_time,omitempty" gorm:"not null"`
}

// HelpPhoto is an image attached to a question or a recommendation, stored as the (already downscaled) bytes.
type HelpPhoto struct {
	Id          string              `json:"id,omitempty" gorm:"primaryKey"`
	AccountId   string              `json:"account_id,omitempty" gorm:"not null;index"`
	Account     *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ContentType string              `json:"content_type,omitempty" gorm:"not null"`
	Data        []byte              `json:"-" gorm:"not null"`
	CreateTime  time.Time           `json:"create_time,omitempty" gorm:"not null"`
}

// PhotoMaxBytes caps an uploaded photo; clients downscale to ~1280 px JPEG first.
const PhotoMaxBytes = 1024 * 1024

// CategoryFollow: the account wants a push for every new question in Category.
type CategoryFollow struct {
	AccountId  string              `json:"account_id,omitempty" gorm:"primaryKey"`
	Account    *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Category   string              `json:"category,omitempty" gorm:"primaryKey"`
	CreateTime time.Time           `json:"create_time,omitempty" gorm:"not null"`
}

// HelpReply is a "Polecam" answer: what the author recommends (a product, a specialist, a place),
// optionally with a link and/or a nanny/tutor profile from the directory.
type HelpReply struct {
	Id          string              `json:"id,omitempty" gorm:"primaryKey"`
	RequestId   string              `json:"request_id,omitempty" gorm:"not null;index"`
	Request     *HelpRequest        `json:"-" gorm:"foreignKey:RequestId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId   string              `json:"account_id,omitempty" gorm:"not null;index"`
	Account     *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Text        string              `json:"text,omitempty" gorm:"not null"`
	Url         string              `json:"url,omitempty"`
	CaregiverId string              `json:"caregiver_id,omitempty"`
	// Optional structured details of the recommendation.
	PhotoId     string    `json:"photo_id,omitempty"`
	ProductName string    `json:"product_name,omitempty"`
	Price       string    `json:"price,omitempty"`
	Rating      int       `json:"rating,omitempty"`
	UsedSince   string    `json:"used_since,omitempty"`
	CreateTime  time.Time `json:"create_time,omitempty" gorm:"not null"`
}

// HelpReplyVote is a "Też polecam" (+1): agreement with somebody else's recommendation. One per (reply, account).
type HelpReplyVote struct {
	ReplyId    string              `json:"reply_id,omitempty" gorm:"primaryKey"`
	Reply      *HelpReply          `json:"-" gorm:"foreignKey:ReplyId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	AccountId  string              `json:"account_id,omitempty" gorm:"primaryKey"`
	Account    *accountspg.Account `json:"-" gorm:"foreignKey:AccountId;references:Id;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	CreateTime time.Time           `json:"create_time,omitempty" gorm:"not null"`
}

// ---- API shapes ----

type RequestItem struct {
	HelpRequest
	AuthorName string `json:"author_name"`
	ReplyCount int64  `json:"reply_count"`
	IsMine     bool   `json:"is_mine"`
	// Preview of the answer shown on the list card: the best one when marked, else the most backed one.
	PreviewText   string `json:"preview_text,omitempty"`
	PreviewAuthor string `json:"preview_author,omitempty"`
	// UnseenReplies is, for the author, how many replies arrived since they last opened the question.
	UnseenReplies int64 `json:"unseen_replies,omitempty"`
}

type ListResponse struct {
	Requests []RequestItem `json:"requests"`
}

type ReplyItem struct {
	HelpReply
	AuthorName    string `json:"author_name"`
	IsMine        bool   `json:"is_mine"`
	CaregiverName string `json:"caregiver_name,omitempty"`
	CaregiverType int    `json:"caregiver_type,omitempty"`
	VoteCount     int64  `json:"vote_count"`
	VotedByMe     bool   `json:"voted_by_me"`
	IsBest        bool   `json:"is_best"`
}

type DetailResponse struct {
	Request RequestItem `json:"request"`
	Replies []ReplyItem `json:"replies"`
}

type CreateRequest struct {
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	City        string `json:"city"`
	PhotoId     string `json:"photo_id"`
}

type ReplyRequest struct {
	RequestId   string `json:"request_id"`
	Text        string `json:"text"`
	Url         string `json:"url"`
	CaregiverId string `json:"caregiver_id"`
	PhotoId     string `json:"photo_id"`
	ProductName string `json:"product_name"`
	Price       string `json:"price"`
	Rating      int    `json:"rating"`
	UsedSince   string `json:"used_since"`
}

type PhotoResponse struct {
	Id string `json:"id"`
}

type FollowsResponse struct {
	Categories []string `json:"categories"`
}

// ActivityItem: one of the viewer's questions that got replies they have not seen yet.
type ActivityItem struct {
	RequestId     string `json:"request_id"`
	Title         string `json:"title"`
	UnseenReplies int64  `json:"unseen_replies"`
}

type ActivityResponse struct {
	Items []ActivityItem `json:"items"`
}

// StatsResponse: how much the viewer has helped others (best answers + "Też polecam" received).
type StatsResponse struct {
	Replies       int64 `json:"replies"`
	BestCount     int64 `json:"best_count"`
	VotesReceived int64 `json:"votes_received"`
	Helped        int64 `json:"helped"`
}

type CategoriesResponse struct {
	Categories []string `json:"categories"`
}
