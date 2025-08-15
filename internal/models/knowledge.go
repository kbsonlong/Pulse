package models

import (
	"time"
	"errors"
	"strings"
	"encoding/json"
)

// KnowledgeType 知识类型
type KnowledgeType string

const (
	KnowledgeTypeArticle    KnowledgeType = "article"    // 文章
	KnowledgeTypeFAQ        KnowledgeType = "faq"        // 常见问题
	KnowledgeTypeTutorial   KnowledgeType = "tutorial"   // 教程
	KnowledgeTypeRunbook    KnowledgeType = "runbook"    // 运维手册
	KnowledgeTypeProcedure  KnowledgeType = "procedure"  // 操作流程
	KnowledgeTypeReference  KnowledgeType = "reference"  // 参考文档
	KnowledgeTypeTroubleshooting KnowledgeType = "troubleshooting" // 故障排除
	KnowledgeTypeTemplate   KnowledgeType = "template"   // 模板
)

// KnowledgeStatus 知识状态
type KnowledgeStatus string

const (
	KnowledgeStatusDraft     KnowledgeStatus = "draft"     // 草稿
	KnowledgeStatusReview    KnowledgeStatus = "review"    // 审核中
	KnowledgeStatusPublished KnowledgeStatus = "published" // 已发布
	KnowledgeStatusArchived  KnowledgeStatus = "archived"  // 已归档
	KnowledgeStatusExpired   KnowledgeStatus = "expired"   // 已过期
)

// KnowledgeVisibility 知识可见性
type KnowledgeVisibility string

const (
	KnowledgeVisibilityPublic   KnowledgeVisibility = "public"   // 公开
	KnowledgeVisibilityInternal KnowledgeVisibility = "internal" // 内部
	KnowledgeVisibilityPrivate  KnowledgeVisibility = "private"  // 私有
	KnowledgeVisibilityTeam     KnowledgeVisibility = "team"     // 团队
)

// KnowledgeFormat 知识格式
type KnowledgeFormat string

const (
	KnowledgeFormatMarkdown KnowledgeFormat = "markdown" // Markdown
	KnowledgeFormatHTML     KnowledgeFormat = "html"     // HTML
	KnowledgeFormatText     KnowledgeFormat = "text"     // 纯文本
	KnowledgeFormatJSON     KnowledgeFormat = "json"     // JSON
	KnowledgeFormatYAML     KnowledgeFormat = "yaml"     // YAML
)

// KnowledgeCategory 知识分类
type KnowledgeCategory struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	ParentID    *string   `json:"parent_id,omitempty" db:"parent_id"`
	Path        string    `json:"path" db:"path"`
	Level       int       `json:"level" db:"level"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	Icon        *string   `json:"icon,omitempty" db:"icon"`
	Color       *string   `json:"color,omitempty" db:"color"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// KnowledgeTag 知识标签
type KnowledgeTag struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	Color       *string   `json:"color,omitempty" db:"color"`
	UsageCount  int64     `json:"usage_count" db:"usage_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// KnowledgeAttachment 知识附件
type KnowledgeAttachment struct {
	ID          string    `json:"id" db:"id"`
	KnowledgeID string    `json:"knowledge_id" db:"knowledge_id"`
	FileName    string    `json:"file_name" db:"file_name"`
	FileSize    int64     `json:"file_size" db:"file_size"`
	FileType    string    `json:"file_type" db:"file_type"`
	FilePath    string    `json:"file_path" db:"file_path"`
	MimeType    string    `json:"mime_type" db:"mime_type"`
	Checksum    string    `json:"checksum" db:"checksum"`
	UploadBy    string    `json:"upload_by" db:"upload_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// KnowledgeVersion 知识版本
type KnowledgeVersion struct {
	ID          string    `json:"id" db:"id"`
	KnowledgeID string    `json:"knowledge_id" db:"knowledge_id"`
	Version     string    `json:"version" db:"version"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	ChangeLog   *string   `json:"change_log,omitempty" db:"change_log"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// KnowledgeMetrics 知识指标
type KnowledgeMetrics struct {
	ViewCount     int64     `json:"view_count"`
	LikeCount     int64     `json:"like_count"`
	DislikeCount  int64     `json:"dislike_count"`
	ShareCount    int64     `json:"share_count"`
	CommentCount  int64     `json:"comment_count"`
	DownloadCount int64     `json:"download_count"`
	Rating        *float64  `json:"rating,omitempty"`
	RatingCount   int64     `json:"rating_count"`
	LastViewedAt  *time.Time `json:"last_viewed_at,omitempty"`
}

// KnowledgeArticle 知识文章（Knowledge的别名）
type KnowledgeArticle = Knowledge

// Knowledge 知识模型
type Knowledge struct {
	ID           string               `json:"id" db:"id"`
	Title        string               `json:"title" db:"title"`
	Slug         string               `json:"slug" db:"slug"`
	Summary      *string              `json:"summary,omitempty" db:"summary"`
	Content      string               `json:"content" db:"content"`
	Type         KnowledgeType        `json:"type" db:"type"`
	Status       KnowledgeStatus      `json:"status" db:"status"`
	Visibility   KnowledgeVisibility  `json:"visibility" db:"visibility"`
	Format       KnowledgeFormat      `json:"format" db:"format"`
	CategoryID   *string              `json:"category_id,omitempty" db:"category_id"`
	Category     *KnowledgeCategory   `json:"category,omitempty" db:"-"`
	Tags         []string             `json:"tags" db:"tags"`
	Keywords     []string             `json:"keywords" db:"keywords"`
	Language     string               `json:"language" db:"language"`
	Version      string               `json:"version" db:"version"`
	AuthorID     string               `json:"author_id" db:"author_id"`
	AuthorName   string               `json:"author_name" db:"author_name"`
	ReviewerID   *string              `json:"reviewer_id,omitempty" db:"reviewer_id"`
	ReviewerName *string              `json:"reviewer_name,omitempty" db:"reviewer_name"`
	TeamID       *string              `json:"team_id,omitempty" db:"team_id"`
	TeamName     *string              `json:"team_name,omitempty" db:"team_name"`
	Priority     int                  `json:"priority" db:"priority"`
	SortOrder    int                  `json:"sort_order" db:"sort_order"`
	IsFeatured   bool                 `json:"is_featured" db:"is_featured"`
	IsTemplate   bool                 `json:"is_template" db:"is_template"`
	TemplateData map[string]interface{} `json:"template_data,omitempty" db:"template_data"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Metrics      *KnowledgeMetrics    `json:"metrics,omitempty" db:"metrics"`
	ViewCount     int64                `json:"view_count" db:"view_count"`
	LikeCount     int64                `json:"like_count" db:"like_count"`
	DislikeCount  int64                `json:"dislike_count" db:"dislike_count"`
	ShareCount    int64                `json:"share_count" db:"share_count"`
	DownloadCount int64                `json:"download_count" db:"download_count"`
	Rating        *float64             `json:"rating,omitempty" db:"rating"`
	RatingCount   int64                `json:"rating_count" db:"rating_count"`
	Featured      bool                 `json:"featured" db:"featured"`
	RelatedIDs    []string             `json:"related_ids" db:"related_ids"`
	ExpiresAt    *time.Time           `json:"expires_at,omitempty" db:"expires_at"`
	PublishedAt  *time.Time           `json:"published_at,omitempty" db:"published_at"`
	ArchivedAt   *time.Time           `json:"archived_at,omitempty" db:"archived_at"`
	ReviewedAt   *time.Time           `json:"reviewed_at,omitempty" db:"reviewed_at"`
	LastEditedAt *time.Time           `json:"last_edited_at,omitempty" db:"last_edited_at"`
	CreatedAt    time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time           `json:"deleted_at,omitempty" db:"deleted_at"`
}

// KnowledgeCreateRequest 创建知识请求
type KnowledgeCreateRequest struct {
	Title        string                 `json:"title" binding:"required,min=1,max=200"`
	Slug         *string                `json:"slug,omitempty"`
	Summary      *string                `json:"summary,omitempty" binding:"omitempty,max=500"`
	Content      string                 `json:"content" binding:"required,min=1"`
	Type         KnowledgeType          `json:"type" binding:"required"`
	Visibility   KnowledgeVisibility    `json:"visibility" binding:"required"`
	Format       KnowledgeFormat        `json:"format" binding:"required"`
	CategoryID   *string                `json:"category_id,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Keywords     []string               `json:"keywords,omitempty"`
	Language     *string                `json:"language,omitempty"`
	TeamID       *string                `json:"team_id,omitempty"`
	Priority     *int                   `json:"priority,omitempty"`
	IsFeatured   *bool                  `json:"is_featured,omitempty"`
	IsTemplate   *bool                  `json:"is_template,omitempty"`
	TemplateData *map[string]interface{} `json:"template_data,omitempty"`
	Metadata     *map[string]interface{} `json:"metadata,omitempty"`
	RelatedIDs   []string               `json:"related_ids,omitempty"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
}

// KnowledgeUpdateRequest 更新知识请求
type KnowledgeUpdateRequest struct {
	Title        *string                `json:"title,omitempty" binding:"omitempty,min=1,max=200"`
	Slug         *string                `json:"slug,omitempty"`
	Summary      *string                `json:"summary,omitempty" binding:"omitempty,max=500"`
	Content      *string                `json:"content,omitempty" binding:"omitempty,min=1"`
	Status       *KnowledgeStatus       `json:"status,omitempty"`
	Visibility   *KnowledgeVisibility   `json:"visibility,omitempty"`
	Format       *KnowledgeFormat       `json:"format,omitempty"`
	CategoryID   *string                `json:"category_id,omitempty"`
	Tags         *[]string              `json:"tags,omitempty"`
	Keywords     *[]string              `json:"keywords,omitempty"`
	Language     *string                `json:"language,omitempty"`
	TeamID       *string                `json:"team_id,omitempty"`
	Priority     *int                   `json:"priority,omitempty"`
	IsFeatured   *bool                  `json:"is_featured,omitempty"`
	IsTemplate   *bool                  `json:"is_template,omitempty"`
	TemplateData *map[string]interface{} `json:"template_data,omitempty"`
	Metadata     *map[string]interface{} `json:"metadata,omitempty"`
	RelatedIDs   *[]string              `json:"related_ids,omitempty"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	ChangeLog    *string                `json:"change_log,omitempty"`
}

// KnowledgePublishRequest 发布知识请求
type KnowledgePublishRequest struct {
	Comment *string `json:"comment,omitempty"`
}

// KnowledgeReviewRequest 审核知识请求
type KnowledgeReviewRequest struct {
	Approved bool    `json:"approved"`
	Comment  *string `json:"comment,omitempty"`
}

// KnowledgeRatingRequest 评分知识请求
type KnowledgeRatingRequest struct {
	Rating  float64 `json:"rating" binding:"required,min=1,max=5"`
	Comment *string `json:"comment,omitempty"`
}

// KnowledgeFilter 知识查询过滤器
type KnowledgeFilter struct {
	Type         *KnowledgeType       `json:"type,omitempty"`
	Status       *KnowledgeStatus     `json:"status,omitempty"`
	Visibility   *KnowledgeVisibility `json:"visibility,omitempty"`
	Format       *KnowledgeFormat     `json:"format,omitempty"`
	CategoryID   *string              `json:"category_id,omitempty"`
	Keyword      *string              `json:"keyword,omitempty"` // 搜索标题、内容
	Tags         []string             `json:"tags,omitempty"`
	Keywords     []string             `json:"keywords,omitempty"`
	Language     *string              `json:"language,omitempty"`
	AuthorID     *string              `json:"author_id,omitempty"`
	TeamID       *string              `json:"team_id,omitempty"`
	IsFeatured   *bool                `json:"is_featured,omitempty"`
	IsTemplate   *bool                `json:"is_template,omitempty"`
	CreatedStart *time.Time           `json:"created_start,omitempty"`
	CreatedEnd   *time.Time           `json:"created_end,omitempty"`
	UpdatedStart *time.Time           `json:"updated_start,omitempty"`
	UpdatedEnd   *time.Time           `json:"updated_end,omitempty"`
	Expired      *bool                `json:"expired,omitempty"`
	Page         int                  `json:"page" binding:"min=1"`
	PageSize     int                  `json:"page_size" binding:"min=1,max=100"`
	SortBy       *string              `json:"sort_by,omitempty"`
	SortOrder    *string              `json:"sort_order,omitempty"` // asc, desc
}

// KnowledgeList 知识列表响应
type KnowledgeList struct {
	Knowledge  []*Knowledge `json:"knowledge"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

// KnowledgeStats 知识统计
type KnowledgeStats struct {
	Total          int64                          `json:"total"`
	ByType         map[KnowledgeType]int64        `json:"by_type"`
	ByStatus       map[KnowledgeStatus]int64      `json:"by_status"`
	ByVisibility   map[KnowledgeVisibility]int64  `json:"by_visibility"`
	ByFormat       map[KnowledgeFormat]int64      `json:"by_format"`
	PublishedCount int64                          `json:"published_count"`
	DraftCount     int64                          `json:"draft_count"`
	FeaturedCount  int64                          `json:"featured_count"`
	TemplateCount  int64                          `json:"template_count"`
	TotalViews     int64                          `json:"total_views"`
	TotalLikes     int64                          `json:"total_likes"`
	AvgRating      float64                        `json:"avg_rating"`
}

// KnowledgeSearchResult 知识搜索结果
type KnowledgeSearchResult struct {
	Knowledge  []*Knowledge `json:"knowledge"`
	Total      int64        `json:"total"`
	Query      string       `json:"query"`
	TookMs     int64        `json:"took_ms"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
	Facets     map[string]interface{} `json:"facets,omitempty"`
}

// 验证方法

// Validate 验证知识数据
func (k *Knowledge) Validate() error {
	if strings.TrimSpace(k.Title) == "" {
		return errors.New("知识标题不能为空")
	}
	
	if len(k.Title) > 200 {
		return errors.New("知识标题长度不能超过200个字符")
	}
	
	if strings.TrimSpace(k.Content) == "" {
		return errors.New("知识内容不能为空")
	}
	
	if !k.Type.IsValid() {
		return errors.New("无效的知识类型")
	}
	
	if !k.Status.IsValid() {
		return errors.New("无效的知识状态")
	}
	
	if !k.Visibility.IsValid() {
		return errors.New("无效的知识可见性")
	}
	
	if !k.Format.IsValid() {
		return errors.New("无效的知识格式")
	}
	
	if strings.TrimSpace(k.AuthorID) == "" {
		return errors.New("作者不能为空")
	}
	
	if strings.TrimSpace(k.Language) == "" {
		k.Language = "zh-CN" // 默认中文
	}
	
	return nil
}

// IsValid 检查知识类型是否有效
func (t KnowledgeType) IsValid() bool {
	switch t {
	case KnowledgeTypeArticle, KnowledgeTypeFAQ, KnowledgeTypeTutorial,
		 KnowledgeTypeRunbook, KnowledgeTypeProcedure, KnowledgeTypeReference,
		 KnowledgeTypeTroubleshooting, KnowledgeTypeTemplate:
		return true
	default:
		return false
	}
}

// IsValid 检查知识状态是否有效
func (s KnowledgeStatus) IsValid() bool {
	switch s {
	case KnowledgeStatusDraft, KnowledgeStatusReview, KnowledgeStatusPublished,
		 KnowledgeStatusArchived, KnowledgeStatusExpired:
		return true
	default:
		return false
	}
}

// IsValid 检查知识可见性是否有效
func (v KnowledgeVisibility) IsValid() bool {
	switch v {
	case KnowledgeVisibilityPublic, KnowledgeVisibilityInternal,
		 KnowledgeVisibilityPrivate, KnowledgeVisibilityTeam:
		return true
	default:
		return false
	}
}

// IsValid 检查知识格式是否有效
func (f KnowledgeFormat) IsValid() bool {
	switch f {
	case KnowledgeFormatMarkdown, KnowledgeFormatHTML, KnowledgeFormatText,
		 KnowledgeFormatJSON, KnowledgeFormatYAML:
		return true
	default:
		return false
	}
}

// Validate 验证创建知识请求
func (req *KnowledgeCreateRequest) Validate() error {
	if strings.TrimSpace(req.Title) == "" {
		return errors.New("知识标题不能为空")
	}
	
	if len(req.Title) > 200 {
		return errors.New("知识标题长度不能超过200个字符")
	}
	
	if strings.TrimSpace(req.Content) == "" {
		return errors.New("知识内容不能为空")
	}
	
	if !req.Type.IsValid() {
		return errors.New("无效的知识类型")
	}
	
	if !req.Visibility.IsValid() {
		return errors.New("无效的知识可见性")
	}
	
	if !req.Format.IsValid() {
		return errors.New("无效的知识格式")
	}
	
	if req.Summary != nil && len(*req.Summary) > 500 {
		return errors.New("知识摘要长度不能超过500个字符")
	}
	
	return nil
}

// Validate 验证知识评分请求
func (req *KnowledgeRatingRequest) Validate() error {
	if req.Rating < 1 || req.Rating > 5 {
		return errors.New("评分必须在1-5之间")
	}
	
	return nil
}

// 辅助方法

// IsPublished 检查知识是否已发布
func (k *Knowledge) IsPublished() bool {
	return k.Status == KnowledgeStatusPublished
}

// IsDraft 检查知识是否为草稿
func (k *Knowledge) IsDraft() bool {
	return k.Status == KnowledgeStatusDraft
}

// IsArchived 检查知识是否已归档
func (k *Knowledge) IsArchived() bool {
	return k.Status == KnowledgeStatusArchived
}

// IsExpired 检查知识是否已过期
func (k *Knowledge) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// IsPublic 检查知识是否公开
func (k *Knowledge) IsPublic() bool {
	return k.Visibility == KnowledgeVisibilityPublic
}

// IsInternal 检查知识是否内部可见
func (k *Knowledge) IsInternal() bool {
	return k.Visibility == KnowledgeVisibilityInternal
}

// IsPrivate 检查知识是否私有
func (k *Knowledge) IsPrivate() bool {
	return k.Visibility == KnowledgeVisibilityPrivate
}

// CanView 检查用户是否可以查看知识
func (k *Knowledge) CanView(userID string, teamID *string) bool {
	// 作者总是可以查看
	if k.AuthorID == userID {
		return true
	}
	
	// 根据可见性检查
	switch k.Visibility {
	case KnowledgeVisibilityPublic:
		return k.IsPublished()
	case KnowledgeVisibilityInternal:
		return k.IsPublished()
	case KnowledgeVisibilityTeam:
		return k.IsPublished() && teamID != nil && k.TeamID != nil && *teamID == *k.TeamID
	case KnowledgeVisibilityPrivate:
		return false
	default:
		return false
	}
}

// CanEdit 检查用户是否可以编辑知识
func (k *Knowledge) CanEdit(userID string, teamID *string) bool {
	// 作者可以编辑
	if k.AuthorID == userID {
		return true
	}
	
	// 团队成员可以编辑团队知识
	if k.TeamID != nil && teamID != nil && *k.TeamID == *teamID {
		return true
	}
	
	return false
}

// GetWordCount 获取字数统计
func (k *Knowledge) GetWordCount() int {
	return len(strings.Fields(k.Content))
}

// GetReadingTime 获取预估阅读时间（分钟）
func (k *Knowledge) GetReadingTime() int {
	wordCount := k.GetWordCount()
	// 假设每分钟阅读200个单词
	readingTime := wordCount / 200
	if readingTime < 1 {
		return 1
	}
	return readingTime
}

// GenerateSlug 生成URL友好的slug
func (k *Knowledge) GenerateSlug() string {
	if k.Slug != "" {
		return k.Slug
	}
	
	// 简单的slug生成逻辑
	slug := strings.ToLower(k.Title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	
	// 移除特殊字符
	allowed := "abcdefghijklmnopqrstuvwxyz0123456789-"
	var result strings.Builder
	for _, char := range slug {
		if strings.ContainsRune(allowed, char) {
			result.WriteRune(char)
		}
	}
	
	return result.String()
}

// MarshalTags 序列化标签为JSON
func (k *Knowledge) MarshalTags() ([]byte, error) {
	if k.Tags == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(k.Tags)
}

// UnmarshalTags 反序列化标签从JSON
func (k *Knowledge) UnmarshalTags(data []byte) error {
	return json.Unmarshal(data, &k.Tags)
}

// MarshalKeywords 序列化关键词为JSON
func (k *Knowledge) MarshalKeywords() ([]byte, error) {
	if k.Keywords == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(k.Keywords)
}

// UnmarshalKeywords 反序列化关键词从JSON
func (k *Knowledge) UnmarshalKeywords(data []byte) error {
	return json.Unmarshal(data, &k.Keywords)
}

// MarshalTemplateData 序列化模板数据为JSON
func (k *Knowledge) MarshalTemplateData() ([]byte, error) {
	if k.TemplateData == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(k.TemplateData)
}

// UnmarshalTemplateData 反序列化模板数据从JSON
func (k *Knowledge) UnmarshalTemplateData(data []byte) error {
	return json.Unmarshal(data, &k.TemplateData)
}

// MarshalMetadata 序列化元数据为JSON
func (k *Knowledge) MarshalMetadata() ([]byte, error) {
	if k.Metadata == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(k.Metadata)
}

// UnmarshalMetadata 反序列化元数据从JSON
func (k *Knowledge) UnmarshalMetadata(data []byte) error {
	return json.Unmarshal(data, &k.Metadata)
}

// MarshalMetrics 序列化指标为JSON
func (k *Knowledge) MarshalMetrics() ([]byte, error) {
	if k.Metrics == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(k.Metrics)
}

// UnmarshalMetrics 反序列化指标从JSON
func (k *Knowledge) UnmarshalMetrics(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		k.Metrics = nil
		return nil
	}
	if k.Metrics == nil {
		k.Metrics = &KnowledgeMetrics{}
	}
	return json.Unmarshal(data, k.Metrics)
}

// MarshalRelatedIDs 序列化相关ID为JSON
func (k *Knowledge) MarshalRelatedIDs() ([]byte, error) {
	if k.RelatedIDs == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(k.RelatedIDs)
}

// UnmarshalRelatedIDs 反序列化相关ID从JSON
func (k *Knowledge) UnmarshalRelatedIDs(data []byte) error {
	return json.Unmarshal(data, &k.RelatedIDs)
}

// GetDisplayName 获取显示名称
func (t KnowledgeType) GetDisplayName() string {
	switch t {
	case KnowledgeTypeArticle:
		return "文章"
	case KnowledgeTypeFAQ:
		return "常见问题"
	case KnowledgeTypeTutorial:
		return "教程"
	case KnowledgeTypeRunbook:
		return "运维手册"
	case KnowledgeTypeProcedure:
		return "操作流程"
	case KnowledgeTypeReference:
		return "参考文档"
	case KnowledgeTypeTroubleshooting:
		return "故障排除"
	case KnowledgeTypeTemplate:
		return "模板"
	default:
		return string(t)
	}
}

// GetDisplayName 获取显示名称
func (s KnowledgeStatus) GetDisplayName() string {
	switch s {
	case KnowledgeStatusDraft:
		return "草稿"
	case KnowledgeStatusReview:
		return "审核中"
	case KnowledgeStatusPublished:
		return "已发布"
	case KnowledgeStatusArchived:
		return "已归档"
	case KnowledgeStatusExpired:
		return "已过期"
	default:
		return string(s)
	}
}

// GetDisplayName 获取显示名称
func (v KnowledgeVisibility) GetDisplayName() string {
	switch v {
	case KnowledgeVisibilityPublic:
		return "公开"
	case KnowledgeVisibilityInternal:
		return "内部"
	case KnowledgeVisibilityPrivate:
		return "私有"
	case KnowledgeVisibilityTeam:
		return "团队"
	default:
		return string(v)
	}
}