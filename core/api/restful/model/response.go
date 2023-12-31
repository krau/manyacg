package model

type Resp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type RespPictureData struct {
	ID        uint    `json:"id"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	DirectURL string  `json:"direct_url"`
	Width     uint    `json:"width"`
	Height    uint    `json:"height"`
	BlurScore float64 `json:"blur_score"`
	Hash      string  `json:"hash"`
	ArtworkID uint    `json:"artwork_id"`
}

type RespArtworkData struct {
	ID          uint              `json:"id"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	Title       string            `json:"title"`
	Author      string            `json:"author"`
	Description string            `json:"description"`
	Source      string            `json:"source"`
	SourceURL   string            `json:"source_url"`
	Tags        []string          `json:"tags"`
	R18         bool              `json:"r18"`
	Pictures    []RespPictureData `json:"pictures"`
}
