package model

// LoginParams represents login parameters.
type LoginParams struct {
	Email           string `json:"email" binding:"required,email" example:"john.doe@gmail.com"`
	Password        string `json:"password" binding:"required" example:"123456789" format:"password"`
	IsLongLiveToken bool   `json:"is_long_live_token"`
}

// SignUpReqBody represents sign-up parameters.
type SignUpReqBody struct {
	// Organization is organization name.
	Organization string `json:"organization" binding:"required"`

	// Name represents user full name.
	Name string `json:"name" binding:"required"`

	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`

	// WithDemoData represents boolean value specifying whether to create demo
	// data.
	WithDemoData bool `json:"withDemoData"`
}

// UserInvitationParams represents user sign-up with existing team parameters.
type UserInvitationParams struct {
	TeamID uint `json:"teamId" binding:"required,gt=0" example:"1"`

	// Name represents user full name.
	Name string `json:"name" binding:"required" example:"Ifrasoft"`

	Email    string `json:"email" binding:"required" example:"john.doe@gmail.com"`
	Password string `json:"password" binding:"required" example:"123456789"`
	LineID   string `json:"lineId" binding:"required" example:"johndoe"`
	RoleID   string `json:"roleId" binding:"required"`
}

// LineSignUpParams represents Line sign-up parameters.
type LineSignUpParams struct {
	// Email represents user email.
	Email    string `json:"email" binding:"required,email" example:"john.doe@gmail.com"`
	Password string `json:"password" binding:"required" example:"123456789" format:"password"`

	// Name represents user full name.
	Name     string `json:"name" binding:"required" example:"John Doe"`
	LineID   string `json:"lineId" binding:"required" example:"johndoe"`
	TeamName string `json:"teamName" binding:"required" example:"Ifrasoft"`

	// RoleID represents roles record ID.
	RoleID string `json:"roleId" binding:"required"`
}

// LineLoginParams represents Line sign-in parameters.
type LineLoginParams struct {
	LineID string `json:"lineId" binding:"required" example:"johndoe" validate:"required"`

	// LineAccessToken represents Line channel access token.
	LineAccessToken string `json:"lineAccessToken" binding:"required" validate:"required"`
}
