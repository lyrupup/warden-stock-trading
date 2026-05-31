package model

// User 用户（多用户预留，V1 单用户模式使用 id=1 的默认用户）。
type User struct {
	BaseModel
	Username     string `gorm:"size:64;not null;uniqueIndex" json:"username"`
	PasswordHash string `gorm:"size:128;not null;default:''" json:"-"`
	Nickname     string `gorm:"size:64;not null;default:''" json:"nickname"`
	Avatar       string `gorm:"size:256;not null;default:''" json:"avatar"`
	Status       int8   `gorm:"not null;default:1" json:"status"`
}

func (User) TableName() string { return "users" }
