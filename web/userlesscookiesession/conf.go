package userlesscookiesession

type Conf struct {
	EncryptionKey string `json:"enckey"`
	ExpireIn      int    `json:"expire_in"` // seconds, -> http.Cookie MaxAge int (Absolute Only)
}
