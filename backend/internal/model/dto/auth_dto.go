package dto

import (
	"regexp"
	"strings"
)

var (
	usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,50}$`)
	passwordRe = regexp.MustCompile(`^[^\s]{8,64}$`)
	emailRe    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	inviteRe   = regexp.MustCompile(`^[A-Za-z0-9_-]{1,32}$`)
)

// --- Register ---

type RegisterSelfReq struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name"`
}

func (r *RegisterSelfReq) Sanitize() {
	r.Username = strings.TrimSpace(strings.ToLower(r.Username))
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.DisplayName = strings.TrimSpace(r.DisplayName)
	if r.DisplayName == "" {
		r.DisplayName = r.Username
	}
}

func (r *RegisterSelfReq) Validate() map[string]string {
	errs := map[string]string{}
	if !usernameRe.MatchString(r.Username) {
		errs["username"] = "username must be 3-50 chars [a-zA-Z0-9_-]"
	}
	if !emailRe.MatchString(r.Email) {
		errs["email"] = "invalid email format"
	}
	if !passwordRe.MatchString(r.Password) {
		errs["password"] = "password must be 8-64 non-space chars"
	}
	if len(r.DisplayName) == 0 || len(r.DisplayName) > 100 {
		errs["display_name"] = "display_name must be 1-100 chars"
	}
	return errs
}

type RegisterInviteReq struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name"`
	InviteCode  string `json:"invite_code" binding:"required"`
}

func (r *RegisterInviteReq) Sanitize() {
	r.Username = strings.TrimSpace(strings.ToLower(r.Username))
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.DisplayName = strings.TrimSpace(r.DisplayName)
	if r.DisplayName == "" {
		r.DisplayName = r.Username
	}
	r.InviteCode = strings.TrimSpace(r.InviteCode)
}

func (r *RegisterInviteReq) Validate() map[string]string {
	errs := map[string]string{}
	if !usernameRe.MatchString(r.Username) {
		errs["username"] = "username must be 3-50 chars [a-zA-Z0-9_-]"
	}
	if !emailRe.MatchString(r.Email) {
		errs["email"] = "invalid email format"
	}
	if !passwordRe.MatchString(r.Password) {
		errs["password"] = "password must be 8-64 non-space chars"
	}
	if len(r.DisplayName) == 0 || len(r.DisplayName) > 100 {
		errs["display_name"] = "display_name must be 1-100 chars"
	}
	if !inviteRe.MatchString(r.InviteCode) {
		errs["invite_code"] = "invite_code must be 1-32 chars [A-Za-z0-9_-]"
	}
	return errs
}

// --- Login ---

type LoginReq struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

func (r *LoginReq) Sanitize() {
	r.UsernameOrEmail = strings.TrimSpace(r.UsernameOrEmail)
}

func (r *LoginReq) Validate() map[string]string {
	errs := map[string]string{}
	if len(r.UsernameOrEmail) == 0 || len(r.UsernameOrEmail) > 150 {
		errs["username_or_email"] = "required, max 150 chars"
	}
	if len(r.Password) == 0 || len(r.Password) > 255 {
		errs["password"] = "required, max 255 chars"
	}
	return errs
}

// --- Refresh / Logout ---

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (r *RefreshReq) Sanitize() { r.RefreshToken = strings.TrimSpace(r.RefreshToken) }
func (r *RefreshReq) Validate() map[string]string {
	if len(r.RefreshToken) == 0 {
		return map[string]string{"refresh_token": "required"}
	}
	return nil
}

type LogoutReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (r *LogoutReq) Sanitize() { r.RefreshToken = strings.TrimSpace(r.RefreshToken) }
func (r *LogoutReq) Validate() map[string]string {
	return nil
}

// --- ValidateInvite ---

type ValidateInviteReq struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

func (r *ValidateInviteReq) Sanitize() { r.InviteCode = strings.TrimSpace(r.InviteCode) }
func (r *ValidateInviteReq) Validate() map[string]string {
	if !inviteRe.MatchString(r.InviteCode) {
		return map[string]string{"invite_code": "must be 1-32 chars [A-Za-z0-9_-]"}
	}
	return nil
}
