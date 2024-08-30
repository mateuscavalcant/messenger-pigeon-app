package model

type User struct {
	ID              int    `json:"id"`
	Username        string `json:"username" binding:"required, min=4,max=32"`
	Name            string `json:"name" binding:"required, min=1,max=70"`
	Icon            []byte `json:"icon"`
	Bio             string `json:"bio" binding:"required, max=70"`
	Email           string `json:"email" binding:"required, email"`
	Password        string `json:"password" binding:"required, min=8, max=16"`
	ConfirmPassword string `json:"cpassword" binding:"required"`
}

type UserMessage struct {
	MessageSession bool   `json:"messagesession"`
	MessageID      int    `json:"post-id"`
	MessageUserID  int    `json:"post-user-id"`
	UserID         int    `json:"user-id"`
	Content        string `json:"content"`
	Icon           []byte `json:"icon"`
	IconBase64     string `json:"iconbase64"`
	CreatedBy      string `json:"createdby"`
	Name           string `json:"createdbyname"`
	MessageBy      int    `json:"message-by"`
	MessageTo      int    `json:"message-to"`
	CreatedAt      string `json:"hourminute"`
}
