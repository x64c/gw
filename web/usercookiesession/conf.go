package usercookiesession

type ExpireMode string

const (
	ExpireAbsolute ExpireMode = "absolute"
	ExpireSliding  ExpireMode = "sliding" // sliding expiration
)

type Conf struct {
	EncryptionKey      string     `json:"enckey"`
	ExpireIn           int        `json:"expire_in"` // seconds, -> http.Cookie MaxAge int
	ExpireMode         ExpireMode `json:"expire_mode"`
	ExtendThreshold    int        `json:"extend_threshold"` // seconds. for Sliding
	WithExternalTokens bool       `json:"with_external_tokens"`
	LoginPath          string     `json:"login_path"`       // login page
	MaxCntPerUser      int64      `json:"max_cnt_per_user"` // max# of cookie sessions per user
}
