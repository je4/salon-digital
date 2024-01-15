package salon

import "github.com/je4/zsearch/v2/pkg/translate"

type Work struct {
	Signature    string
	Title        *translate.MultiLangString
	Year         string
	Authors      []string
	Description  *translate.MultiLangString
	ImageUrl     string
	ThumbnailUrl string
	IFrameUrl    string
}
