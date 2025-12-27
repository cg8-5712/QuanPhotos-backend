package i18n

import (
	"sync"
)

// Language codes
const (
	LangZhCN = "zh-CN"
	LangEnUS = "en-US"
)

// DefaultLang is the default language
const DefaultLang = LangZhCN

// Message keys for error messages
const (
	// Auth errors
	MsgUnauthorized      = "error.unauthorized"
	MsgTokenExpired      = "error.token_expired"
	MsgTokenInvalid      = "error.token_invalid"
	MsgForbidden         = "error.forbidden"
	MsgInvalidCredentials = "error.invalid_credentials"

	// Validation errors
	MsgInvalidParams     = "error.invalid_params"
	MsgValidationFailed  = "error.validation_failed"
	MsgFieldRequired     = "error.field_required"
	MsgFieldTooLong      = "error.field_too_long"
	MsgFieldTooShort     = "error.field_too_short"
	MsgInvalidEmail      = "error.invalid_email"
	MsgInvalidPassword   = "error.invalid_password"
	MsgPasswordMismatch  = "error.password_mismatch"

	// Resource errors
	MsgNotFound          = "error.not_found"
	MsgConflict          = "error.conflict"
	MsgDuplicateEntry    = "error.duplicate_entry"
	MsgUserNotFound      = "error.user_not_found"
	MsgPhotoNotFound     = "error.photo_not_found"
	MsgCategoryNotFound  = "error.category_not_found"
	MsgTagNotFound       = "error.tag_not_found"
	MsgTicketNotFound    = "error.ticket_not_found"

	// File errors
	MsgFileTooLarge      = "error.file_too_large"
	MsgInvalidFileType   = "error.invalid_file_type"
	MsgUploadFailed      = "error.upload_failed"

	// Rate limit
	MsgTooManyRequests   = "error.too_many_requests"

	// Server errors
	MsgInternalError     = "error.internal_error"
	MsgServiceUnavailable = "error.service_unavailable"

	// Success messages
	MsgSuccess           = "success.default"
	MsgCreated           = "success.created"
	MsgUpdated           = "success.updated"
	MsgDeleted           = "success.deleted"
	MsgLoginSuccess      = "success.login"
	MsgLogoutSuccess     = "success.logout"
	MsgRegisterSuccess   = "success.register"

	// Photo specific
	MsgPhotoUploaded     = "photo.uploaded"
	MsgPhotoApproved     = "photo.approved"
	MsgPhotoRejected     = "photo.rejected"
	MsgPhotoDeleted      = "photo.deleted"
	MsgAddedToFavorites  = "photo.added_to_favorites"
	MsgRemovedFromFavorites = "photo.removed_from_favorites"
	MsgLiked             = "photo.liked"
	MsgUnliked           = "photo.unliked"

	// Review specific
	MsgReviewPending     = "review.pending"
	MsgReviewApproved    = "review.approved"
	MsgReviewRejected    = "review.rejected"

	// Ticket specific
	MsgTicketCreated     = "ticket.created"
	MsgTicketUpdated     = "ticket.updated"
	MsgTicketResolved    = "ticket.resolved"
	MsgTicketClosed      = "ticket.closed"
)

// I18n manages internationalization
type I18n struct {
	mu       sync.RWMutex
	messages map[string]map[string]string
	lang     string
}

// global instance
var instance *I18n
var once sync.Once

// Init initializes the global I18n instance
func Init() *I18n {
	once.Do(func() {
		instance = &I18n{
			messages: make(map[string]map[string]string),
			lang:     DefaultLang,
		}
		instance.loadMessages()
	})
	return instance
}

// Get returns the global I18n instance
func Get() *I18n {
	if instance == nil {
		return Init()
	}
	return instance
}

// SetLang sets the current language
func (i *I18n) SetLang(lang string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, ok := i.messages[lang]; ok {
		i.lang = lang
	}
}

// GetLang returns the current language
func (i *I18n) GetLang() string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.lang
}

// T translates a message key to the current language
func (i *I18n) T(key string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if msgs, ok := i.messages[i.lang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	// Fallback to default language
	if msgs, ok := i.messages[DefaultLang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	return key
}

// TLang translates a message key to a specific language
func (i *I18n) TLang(lang, key string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if msgs, ok := i.messages[lang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	// Fallback to default language
	if msgs, ok := i.messages[DefaultLang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	return key
}

// T is a convenience function for translating with the global instance
func T(key string) string {
	return Get().T(key)
}

// TLang is a convenience function for translating with a specific language
func TLang(lang, key string) string {
	return Get().TLang(lang, key)
}

// loadMessages loads all message translations
func (i *I18n) loadMessages() {
	i.messages[LangZhCN] = zhCNMessages
	i.messages[LangEnUS] = enUSMessages
}

// SupportedLanguages returns a list of supported languages
func SupportedLanguages() []string {
	return []string{LangZhCN, LangEnUS}
}

// IsSupported checks if a language is supported
func IsSupported(lang string) bool {
	for _, l := range SupportedLanguages() {
		if l == lang {
			return true
		}
	}
	return false
}
