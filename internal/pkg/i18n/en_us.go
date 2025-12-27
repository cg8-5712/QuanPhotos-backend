package i18n

// enUSMessages contains English (US) translations
var enUSMessages = map[string]string{
	// Auth errors
	MsgUnauthorized:       "Unauthorized access",
	MsgTokenExpired:       "Session expired, please login again",
	MsgTokenInvalid:       "Invalid authentication token",
	MsgForbidden:          "You don't have permission to perform this action",
	MsgInvalidCredentials: "Invalid username or password",

	// Validation errors
	MsgInvalidParams:    "Invalid request parameters",
	MsgValidationFailed: "Validation failed",
	MsgFieldRequired:    "This field is required",
	MsgFieldTooLong:     "Field exceeds maximum length",
	MsgFieldTooShort:    "Field is too short",
	MsgInvalidEmail:     "Invalid email format",
	MsgInvalidPassword:  "Invalid password format",
	MsgPasswordMismatch: "Passwords do not match",

	// Resource errors
	MsgNotFound:         "Resource not found",
	MsgConflict:         "Resource conflict",
	MsgDuplicateEntry:   "Entry already exists",
	MsgUserNotFound:     "User not found",
	MsgPhotoNotFound:    "Photo not found",
	MsgCategoryNotFound: "Category not found",
	MsgTagNotFound:      "Tag not found",
	MsgTicketNotFound:   "Ticket not found",

	// File errors
	MsgFileTooLarge:    "File size exceeds limit",
	MsgInvalidFileType: "Unsupported file format",
	MsgUploadFailed:    "File upload failed",

	// Rate limit
	MsgTooManyRequests: "Too many requests, please try again later",

	// Server errors
	MsgInternalError:      "Internal server error",
	MsgServiceUnavailable: "Service temporarily unavailable",

	// Success messages
	MsgSuccess:         "Operation successful",
	MsgCreated:         "Created successfully",
	MsgUpdated:         "Updated successfully",
	MsgDeleted:         "Deleted successfully",
	MsgLoginSuccess:    "Login successful",
	MsgLogoutSuccess:   "Logged out successfully",
	MsgRegisterSuccess: "Registration successful",

	// Photo specific
	MsgPhotoUploaded:        "Photo uploaded successfully",
	MsgPhotoApproved:        "Photo approved",
	MsgPhotoRejected:        "Photo rejected",
	MsgPhotoDeleted:         "Photo deleted",
	MsgAddedToFavorites:     "Added to favorites",
	MsgRemovedFromFavorites: "Removed from favorites",
	MsgLiked:                "Liked",
	MsgUnliked:              "Unliked",

	// Review specific
	MsgReviewPending:  "Photo pending review",
	MsgReviewApproved: "Review approved",
	MsgReviewRejected: "Review rejected",

	// Ticket specific
	MsgTicketCreated:  "Ticket submitted",
	MsgTicketUpdated:  "Ticket updated",
	MsgTicketResolved: "Ticket resolved",
	MsgTicketClosed:   "Ticket closed",
}
