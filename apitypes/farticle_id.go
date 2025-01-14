package apitypes

import (
	"github.com/Ptt-official-app/go-pttbbs/bbs"
	"github.com/Ptt-official-app/go-pttbbs/ptttype"
)

//FArticleID as article-id for frontend.
type FArticleID string

func ToFArticleID(articleID bbs.ArticleID) FArticleID {
	return FArticleID(articleID.ToFilename().String())
}

func (f FArticleID) ToArticleID() bbs.ArticleID {
	filename := &ptttype.Filename_t{}
	copy(filename[:], f[:])

	return bbs.ToArticleID(filename)
}
