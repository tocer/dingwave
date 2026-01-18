package api

import (
	"strconv"

	"dingtalk/internal/database"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

type ConversationSection struct {
	Total int64                    `json:"total"`
	Items []database.Conversation `json:"items"`
}

type ConversationsHomeResponse struct {
	CurrentUser *database.CurrentUser   `json:"current_user"`
	Top         ConversationSection     `json:"top"`
	Single      ConversationSection     `json:"single"`
	Group       ConversationSection     `json:"group"`
}

func (h *Handler) GetConversationsHome(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 5
	}

	var resp ConversationsHomeResponse

	// Current user
	var currentUser database.CurrentUser
	h.db.First(&currentUser)
	resp.CurrentUser = &currentUser

	// Top conversations
	h.db.Model(&database.Conversation{}).Where("is_top = ?", true).Count(&resp.Top.Total)
	h.db.Where("is_top = ?", true).Order("last_message_at DESC").Limit(limit).Find(&resp.Top.Items)

	// Single chats
	h.db.Model(&database.Conversation{}).Where("type = ?", database.ConversationTypeSingle).Count(&resp.Single.Total)
	h.db.Where("type = ?", database.ConversationTypeSingle).Order("last_message_at DESC").Limit(limit).Find(&resp.Single.Items)

	// Group chats
	h.db.Model(&database.Conversation{}).Where("type = ?", database.ConversationTypeGroup).Count(&resp.Group.Total)
	h.db.Where("type = ?", database.ConversationTypeGroup).Order("last_message_at DESC").Limit(limit).Find(&resp.Group.Items)

	return Success(c, resp)
}

type ConversationsResponse struct {
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
	Items []database.Conversation `json:"items"`
}

func (h *Handler) GetConversations(c echo.Context) error {
	convType, _ := strconv.Atoi(c.QueryParam("type"))
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	size, _ := strconv.Atoi(c.QueryParam("size"))
	if size <= 0 {
		size = 20
	}
	order := c.QueryParam("order")
	if order == "" {
		order = "time"
	}

	var resp ConversationsResponse
	resp.Page = page
	resp.Size = size
	offset := (page - 1) * size

	query := h.db.Model(&database.Conversation{})
	if convType == 0 {
		query = query.Where("is_top = ?", true)
	} else {
		query = query.Where("type = ?", convType)
	}

	query.Count(&resp.Total)

	if order == "count" {
		query = query.Order("message_count DESC")
	} else {
		query = query.Order("last_message_at DESC")
	}

	query.Limit(size).Offset(offset).Find(&resp.Items)
	return Success(c, resp)
}

type UsersResponse struct {
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
	Items []database.User `json:"items"`
}

func (h *Handler) GetUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	size, _ := strconv.Atoi(c.QueryParam("size"))
	if size <= 0 {
		size = 50
	}

	var resp UsersResponse
	resp.Page = page
	resp.Size = size

	offset := (page - 1) * size

	h.db.Model(&database.User{}).Count(&resp.Total)
	h.db.Limit(size).Offset(offset).Find(&resp.Items)

	return Success(c, resp)
}

func (h *Handler) SearchUsers(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return Error(c, 400, "q is required")
	}
	size, _ := strconv.Atoi(c.QueryParam("size"))
	if size <= 0 {
		size = 20
	}

	var users []database.User
	h.db.Where("nickname LIKE ? OR email LIKE ?", "%"+q+"%", "%"+q+"%").Limit(size).Find(&users)

	return Success(c, users)
}

type MessagesResponse struct {
	HasMore bool               `json:"has_more"`
	Items   []database.Message `json:"items"`
}

func (h *Handler) GetConversationMessages(c echo.Context) error {
	cid := c.Param("cid")
	before, _ := strconv.ParseInt(c.QueryParam("before"), 10, 64)
	after, _ := strconv.ParseInt(c.QueryParam("after"), 10, 64)
	size, _ := strconv.Atoi(c.QueryParam("size"))
	if size <= 0 {
		size = 50
	}

	var resp MessagesResponse
	query := h.db.Where("cid = ?", cid)

	if before > 0 {
		query = query.Where("created_at < ?", before).Order("created_at DESC")
	} else if after > 0 {
		query = query.Where("created_at > ?", after).Order("created_at ASC")
	} else {
		query = query.Order("created_at DESC")
	}

	query.Limit(size + 1).Find(&resp.Items)

	if len(resp.Items) > size {
		resp.HasMore = true
		resp.Items = resp.Items[:size]
	}

	// Populate sender names
	userMap := make(map[int64]string)
	for i := range resp.Items {
		if _, exists := userMap[resp.Items[i].SenderID]; !exists {
			var user database.User
			if err := h.db.First(&user, resp.Items[i].SenderID).Error; err == nil {
				userMap[resp.Items[i].SenderID] = user.Nickname
			}
		}
		resp.Items[i].SenderName = userMap[resp.Items[i].SenderID]
		resp.Items[i].ContentTypeStr = resp.Items[i].ContentType.String()
	}

	if after > 0 {
		for i, j := 0, len(resp.Items)-1; i < j; i, j = i+1, j-1 {
			resp.Items[i], resp.Items[j] = resp.Items[j], resp.Items[i]
		}
	}

	// Reverse to return in ascending order
	for i, j := 0, len(resp.Items)-1; i < j; i, j = i+1, j-1 {
		resp.Items[i], resp.Items[j] = resp.Items[j], resp.Items[i]
	}

	return Success(c, resp)
}

type SearchResultItem struct {
	CID        string                   `json:"cid"`
	Title      string                   `json:"title"`
	Type       database.ConversationType `json:"type"`
	MatchCount int64                    `json:"match_count"`
}

type SearchMessagesResponse struct {
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
	Items []SearchResultItem `json:"items"`
}

func (h *Handler) SearchMessagesGlobal(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return Error(c, 400, "q is required")
	}
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	size, _ := strconv.Atoi(c.QueryParam("size"))
	if size <= 0 {
		size = 20
	}

	var resp SearchMessagesResponse
	resp.Page = page
	resp.Size = size
	resp.Items = []SearchResultItem{}
	offset := (page - 1) * size

	// Count distinct conversations with matches
	h.db.Model(&database.Message{}).
		Where("content_text LIKE ?", "%"+q+"%").
		Distinct("cid").
		Count(&resp.Total)

	// Get aggregated results
	type AggResult struct {
		CID        string `gorm:"column:cid"`
		MatchCount int64  `gorm:"column:match_count"`
	}
	var aggResults []AggResult
	h.db.Model(&database.Message{}).
		Select("cid, COUNT(*) as match_count").
		Where("content_text LIKE ?", "%"+q+"%").
		Group("cid").
		Order("match_count DESC").
		Limit(size).
		Offset(offset).
		Scan(&aggResults)

	// Get conversation info
	for _, agg := range aggResults {
		var conv database.Conversation
		if err := h.db.Where("cid = ?", agg.CID).First(&conv).Error; err == nil {
			resp.Items = append(resp.Items, SearchResultItem{
				CID:        agg.CID,
				Title:      conv.Title,
				Type:       conv.Type,
				MatchCount: agg.MatchCount,
			})
		}
	}

	return Success(c, resp)
}

type SearchConvMessagesResponse struct {
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
	Items []database.Message `json:"items"`
}

func (h *Handler) SearchConversationMessages(c echo.Context) error {
	cid := c.Param("cid")
	q := c.QueryParam("q")
	if q == "" {
		return Error(c, 400, "q is required")
	}
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	size, _ := strconv.Atoi(c.QueryParam("size"))
	if size <= 0 {
		size = 20
	}

	var resp SearchConvMessagesResponse
	resp.Page = page
	resp.Size = size
	resp.Items = []database.Message{}
	offset := (page - 1) * size

	query := h.db.Model(&database.Message{}).Where("cid = ? AND content_text LIKE ?", cid, "%"+q+"%")
	query.Count(&resp.Total)
	query.Order("created_at DESC").Limit(size).Offset(offset).Find(&resp.Items)

	// Populate sender names
	userMap := make(map[int64]string)
	for i := range resp.Items {
		if _, exists := userMap[resp.Items[i].SenderID]; !exists {
			var user database.User
			if err := h.db.First(&user, resp.Items[i].SenderID).Error; err == nil {
				userMap[resp.Items[i].SenderID] = user.Nickname
			}
		}
		resp.Items[i].SenderName = userMap[resp.Items[i].SenderID]
		resp.Items[i].ContentTypeStr = resp.Items[i].ContentType.String()
	}

	return Success(c, resp)
}
