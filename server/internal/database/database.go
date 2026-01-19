package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"dingtalk/internal/logger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func MigrateToMemory(dbPath string) (*gorm.DB, error) {
	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}

	memDB, err := createMemoryDB()
	if err != nil {
		return nil, err
	}

	// 迁移 tbuser_profile_v2
	if err := migrateUsers(db, memDB); err != nil {
		return nil, fmt.Errorf("failed to migrate users: %w", err)
	}

	// 迁移 tbconversation
	if err := migrateConversations(db, memDB); err != nil {
		return nil, fmt.Errorf("failed to migrate conversations: %w", err)
	}

	// 合并与迁移钉钉数据库中的多个 tbmessage 表
	if err := migrateMessages(db, memDB); err != nil {
		return nil, fmt.Errorf("failed to migrate messages: %w", err)
	}

	// 更新消息内容文本
	if err := updateContentText(memDB); err != nil {
		return nil, fmt.Errorf("failed to update content text: %w", err)
	}

	// 更新单聊会话标题
	if err := updateSingleChatTitles(memDB); err != nil {
		return nil, fmt.Errorf("failed to update single chat titles: %w", err)
	}

	// 保存当前用户信息
	if err := saveCurrentUser(memDB); err != nil {
		return nil, fmt.Errorf("failed to save current user: %w", err)
	}

	// 更新会话统计信息
	if err := updateConversationStats(memDB); err != nil {
		return nil, fmt.Errorf("failed to update conversation stats: %w", err)
	}

	// 输出统计信息
	if err := logStatistics(memDB); err != nil {
		logger.Warn("failed to log statistics: %v", err)
	}

	return memDB, nil
}

func migrateUsers(srcDB *sql.DB, destDB *gorm.DB) error {
	var users []User
	rows, err := srcDB.Query("SELECT uid,nick,email FROM tbuser_profile_v2")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Nickname, &user.Email); err != nil {
			return err
		}
		users = append(users, user)
	}

	if err := destDB.Create(&users).Error; err != nil {
		return err
	}

	return nil
}

func migrateConversations(srcDB *sql.DB, destDB *gorm.DB) error {
	var conversations []Conversation
	var top int
	rows, err := srcDB.Query("SELECT cid,type,title,top,lastMid,createAt FROM tbconversation")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.CID, &conv.Type, &conv.Title, &top, &conv.LastMessageID, &conv.CreatedAt); err != nil {
			return err
		}
		if top > 0 {
			conv.IsTop = true
		}
		conversations = append(conversations, conv)
	}

	if err := destDB.Create(&conversations).Error; err != nil {
		return err
	}

	return nil
}

func migrateMessages(srcDB *sql.DB, destDB *gorm.DB) error {
	rows, err := srcDB.Query("SELECT name FROM sqlite_master WHERE type = 'table' AND name LIKE 'tbmsg%';")
	if err != nil {
		return err
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return err
		}
		tableNames = append(tableNames, t)
	}

	const batchSize = 1000
	for _, tableName := range tableNames {
		offset := 0
		for {
			query := fmt.Sprintf("SELECT mid, cid, senderId, contentType, content, createdAt, recallStatus FROM %s LIMIT %d OFFSET %d", tableName, batchSize, offset)
			rows, err := srcDB.Query(query)
			if err != nil {
				return err
			}

			var batch []Message
			for rows.Next() {
				var msg Message
				var recallStatus int
				if err := rows.Scan(&msg.ID, &msg.CID, &msg.SenderID, &msg.ContentType, &msg.ContentJson, &msg.CreatedAt, &recallStatus); err != nil {
					rows.Close()
					return err
				}
				if recallStatus > 0 {
					msg.IsRecall = true
				}
				batch = append(batch, msg)
			}
			rows.Close()

			if len(batch) == 0 {
				break
			}

			if err := destDB.Create(&batch).Error; err != nil {
				return err
			}

			offset += batchSize
		}
	}

	return nil
}

func updateContentText(db *gorm.DB) error {
	const batchSize = 500
	offset := 0

	for {
		var messages []Message
		if err := db.Limit(batchSize).Offset(offset).Find(&messages).Error; err != nil {
			return err
		}

		if len(messages) == 0 {
			break
		}

		for i := range messages {
			messages[i].ContentText = extractContentText(messages[i].ContentType, messages[i].ContentJson)
		}

		if err := db.Save(&messages).Error; err != nil {
			return err
		}

		offset += batchSize
	}

	return nil
}

func updateSingleChatTitles(db *gorm.DB) error {
	currentUserID, err := getCurrentUserID(db)
	if err != nil {
		return err
	}

	var conversations []Conversation
	if err := db.Where("type = ?", ConversationTypeSingle).Find(&conversations).Error; err != nil {
		return err
	}

	for i := range conversations {
		parts := strings.Split(conversations[i].CID, ":")
		if len(parts) != 2 {
			continue
		}

		id1, _ := strconv.ParseInt(parts[0], 10, 64)
		id2, _ := strconv.ParseInt(parts[1], 10, 64)

		var otherUserID int64
		if id1 == currentUserID {
			otherUserID = id2
		} else {
			otherUserID = id1
		}

		var user User
		if err := db.First(&user, otherUserID).Error; err == nil {
			conversations[i].Title = user.Nickname
		}
	}

	return db.Save(&conversations).Error
}

func getCurrentUserID(db *gorm.DB) (int64, error) {
	var conversations []Conversation
	if err := db.Where("type = ?", ConversationTypeSingle).Find(&conversations).Error; err != nil {
		return 0, err
	}

	idCount := make(map[int64]int)
	for _, conv := range conversations {
		parts := strings.Split(conv.CID, ":")
		if len(parts) == 2 {
			if id, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				idCount[id]++
			}
			if id, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				idCount[id]++
			}
		}
	}

	var currentUserID int64
	maxCount := 0
	for id, count := range idCount {
		if count > maxCount {
			maxCount = count
			currentUserID = id
		}
	}

	return currentUserID, nil
}

func saveCurrentUser(db *gorm.DB) error {
	currentUserID, err := getCurrentUserID(db)
	if err != nil {
		return err
	}

	var user User
	if err := db.First(&user, currentUserID).Error; err != nil {
		return err
	}

	currentUser := CurrentUser{
		ID:       user.ID,
		Nickname: user.Nickname,
		Email:    user.Email,
	}

	return db.Create(&currentUser).Error
}

func updateConversationStats(db *gorm.DB) error {
	var conversations []Conversation
	if err := db.Find(&conversations).Error; err != nil {
		return err
	}

	for i := range conversations {
		var count int64
		if err := db.Model(&Message{}).Where("cid = ?", conversations[i].CID).Count(&count).Error; err != nil {
			return err
		}
		conversations[i].MessageCount = count

		var lastMsg Message
		if err := db.Where("cid = ?", conversations[i].CID).Order("created_at DESC").First(&lastMsg).Error; err == nil {
			conversations[i].LastMessageAt = lastMsg.CreatedAt
			conversations[i].LastMessagePreview = lastMsg.ContentText
		}
	}

	return db.Save(&conversations).Error
}

func extractContentText(contentType MessageContentType, contentJson string) string {
	var content map[string]interface{}
	if err := json.Unmarshal([]byte(contentJson), &content); err != nil {
		return ""
	}

	switch contentType {
	case MessageContentTypeText:
		if text, ok := content["text"].(string); ok {
			return text
		}
	case MessageContentTypeImage:
		if filename, ok := content["filename"].(string); ok {
			return "[图片] " + filename
		}
		logger.Warn("Image content fallback to default, contentType=%d", contentType)
		return "[图片]"
	case MessageContentTypeDocument:
		if filename, ok := content["filename"].(string); ok {
			return "[文件] " + filename
		}
		return "[文件]"
	case MessageContentTypeShareLink:
		return extractAttachmentUrl(content, "[链接]")
	case MessageContentTypeLink:
		return extractFromAttachments(content, "b_tl", "[链接]", contentType)
	case MessageContentTypeFileOld, MessageContentTypeFile:
		return extractFromAttachments(content, "f_name", "[文件]", contentType)
	case MessageContentTypeFolder:
		return extractFromAttachments(content, "f_name", "[文件夹]", contentType)
	case MessageContentTypeSticker:
		return "[表情]"
	case MessageContentTypeCard:
		return "[名片]"
	case MessageContentTypeVideo:
		return extractFromAttachments(content, "title", "[视频]", contentType)
	case MessageContentTypeShortVideo:
		return "[短视频]"
	case MessageContentTypeVideoCall:
		return extractFromAttachments(content, "title", "[视频通话]", contentType)
	case MessageContentTypeCalendar:
		return "[日程]"
	case MessageContentTypeVote:
		return "[投票]"
	case MessageContentTypeRobot:
		return extractLastMessageI18n(content, "[群公告]", contentType)
	case MessageContentTypeActionCard:
		return extractLastMessageI18n(content, "[互动卡片]", contentType)
	case MessageContentTypeMiniProgram:
		return extractFromAttachments(content, "desc", "[小程序]", contentType)
	}
	return ""
}

func extractAttachmentUrl(content map[string]interface{}, fallback string) string {
	attachments, ok := content["attachments"].([]interface{})
	if !ok || len(attachments) == 0 {
		return fallback
	}
	att, ok := attachments[0].(map[string]interface{})
	if !ok {
		return fallback
	}
	if url, ok := att["url"].(string); ok && url != "" {
		return url
	}
	return fallback
}

func extractFromAttachments(content map[string]interface{}, field, fallback string, contentType MessageContentType) string {
	attachments, ok := content["attachments"].([]interface{})
	if !ok || len(attachments) == 0 {
		logger.Warn("Content fallback to default, contentType=%d, reason=no attachments", contentType)
		return fallback
	}

	att, ok := attachments[0].(map[string]interface{})
	if !ok {
		logger.Warn("Content fallback to default, contentType=%d, reason=invalid attachment", contentType)
		return fallback
	}

	extStr, ok := att["extension"].(string)
	if !ok {
		logger.Warn("Content fallback to default, contentType=%d, reason=no extension", contentType)
		return fallback
	}

	var ext map[string]interface{}
	if err := json.Unmarshal([]byte(extStr), &ext); err != nil {
		logger.Warn("Content fallback to default, contentType=%d, reason=invalid extension json", contentType)
		return fallback
	}

	if val, ok := ext[field].(string); ok && val != "" {
		return val
	}

	logger.Warn("Content fallback to default, contentType=%d, reason=field '%s' not found", contentType, field)
	return fallback
}

func extractLastMessageI18n(content map[string]interface{}, fallback string, contentType MessageContentType) string {
	attachments, ok := content["attachments"].([]interface{})
	if !ok || len(attachments) == 0 {
		logger.Warn("Content fallback to default, contentType=%d, reason=no attachments", contentType)
		return fallback
	}

	att, ok := attachments[0].(map[string]interface{})
	if !ok {
		logger.Warn("Content fallback to default, contentType=%d, reason=invalid attachment", contentType)
		return fallback
	}

	extStr, ok := att["extension"].(string)
	if !ok {
		logger.Warn("Content fallback to default, contentType=%d, reason=no extension", contentType)
		return fallback
	}

	var ext map[string]interface{}
	if err := json.Unmarshal([]byte(extStr), &ext); err != nil {
		logger.Warn("Content fallback to default, contentType=%d, reason=invalid extension json", contentType)
		return fallback
	}

	if val, ok := ext["interactiveCardLastMessage"].(string); ok && val != "" {
		return val
	}

	i18nStr, ok := ext["LastMessageI18n"].(string)
	if !ok {
		logger.Warn("Content fallback to default, contentType=%d, reason=no LastMessageI18n or interactiveCardLastMessage", contentType)
		return fallback
	}

	var i18n map[string]interface{}
	if err := json.Unmarshal([]byte(i18nStr), &i18n); err != nil {
		logger.Warn("Content fallback to default, contentType=%d, reason=invalid i18n json", contentType)
		return fallback
	}

	if val, ok := i18n["zh_CN"].(string); ok && val != "" {
		return val
	}
	logger.Warn("Content fallback to default, contentType=%d, reason=zh_CN not found", contentType)
	return fallback
}

func openDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func createMemoryDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create memory database: %w", err)
	}

	if err := db.AutoMigrate(
		&Conversation{},
		&User{},
		&CurrentUser{},
		&Message{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return db, nil
}

func logStatistics(db *gorm.DB) error {
	var totalMessages int64
	if err := db.Model(&Message{}).Count(&totalMessages).Error; err != nil {
		return err
	}

	var totalConversations int64
	if err := db.Model(&Conversation{}).Count(&totalConversations).Error; err != nil {
		return err
	}

	var topConversations int64
	if err := db.Model(&Conversation{}).Where("is_top = ?", true).Count(&topConversations).Error; err != nil {
		return err
	}

	var singleChats int64
	if err := db.Model(&Conversation{}).Where("type = ?", ConversationTypeSingle).Count(&singleChats).Error; err != nil {
		return err
	}

	var groupChats int64
	if err := db.Model(&Conversation{}).Where("type = ?", ConversationTypeGroup).Count(&groupChats).Error; err != nil {
		return err
	}

	var totalUsers int64
	if err := db.Model(&User{}).Count(&totalUsers).Error; err != nil {
		return err
	}

	logger.Info("total messages: %d", totalMessages)
	logger.Info("total conversations: %d[top: %d, single: %d, group: %d]", totalConversations, topConversations, singleChats, groupChats)
	logger.Info("total users: %d", totalUsers)

	return nil
}
