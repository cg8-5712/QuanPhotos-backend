package i18n

// zhCNMessages contains Chinese (Simplified) translations
var zhCNMessages = map[string]string{
	// Auth errors
	MsgUnauthorized:       "未授权访问",
	MsgTokenExpired:       "登录已过期，请重新登录",
	MsgTokenInvalid:       "无效的登录凭证",
	MsgForbidden:          "没有权限执行此操作",
	MsgInvalidCredentials: "用户名或密码错误",

	// Validation errors
	MsgInvalidParams:    "请求参数无效",
	MsgValidationFailed: "数据验证失败",
	MsgFieldRequired:    "必填字段不能为空",
	MsgFieldTooLong:     "字段长度超出限制",
	MsgFieldTooShort:    "字段长度不足",
	MsgInvalidEmail:     "邮箱格式不正确",
	MsgInvalidPassword:  "密码格式不正确",
	MsgPasswordMismatch: "两次输入的密码不一致",

	// Resource errors
	MsgNotFound:         "资源不存在",
	MsgConflict:         "资源冲突",
	MsgDuplicateEntry:   "数据已存在",
	MsgUserNotFound:     "用户不存在",
	MsgPhotoNotFound:    "照片不存在",
	MsgCategoryNotFound: "分类不存在",
	MsgTagNotFound:      "标签不存在",
	MsgTicketNotFound:   "工单不存在",

	// File errors
	MsgFileTooLarge:    "文件大小超出限制",
	MsgInvalidFileType: "不支持的文件格式",
	MsgUploadFailed:    "文件上传失败",

	// Rate limit
	MsgTooManyRequests: "请求过于频繁，请稍后再试",

	// Server errors
	MsgInternalError:      "服务器内部错误",
	MsgServiceUnavailable: "服务暂时不可用",

	// Success messages
	MsgSuccess:         "操作成功",
	MsgCreated:         "创建成功",
	MsgUpdated:         "更新成功",
	MsgDeleted:         "删除成功",
	MsgLoginSuccess:    "登录成功",
	MsgLogoutSuccess:   "退出登录成功",
	MsgRegisterSuccess: "注册成功",

	// Photo specific
	MsgPhotoUploaded:        "照片上传成功",
	MsgPhotoApproved:        "照片审核通过",
	MsgPhotoRejected:        "照片审核未通过",
	MsgPhotoDeleted:         "照片已删除",
	MsgAddedToFavorites:     "已添加到收藏",
	MsgRemovedFromFavorites: "已从收藏移除",
	MsgLiked:                "已点赞",
	MsgUnliked:              "已取消点赞",

	// Review specific
	MsgReviewPending:  "照片待审核",
	MsgReviewApproved: "审核已通过",
	MsgReviewRejected: "审核未通过",

	// Ticket specific
	MsgTicketCreated:  "工单已提交",
	MsgTicketUpdated:  "工单已更新",
	MsgTicketResolved: "工单已解决",
	MsgTicketClosed:   "工单已关闭",
}
