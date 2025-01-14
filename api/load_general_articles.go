package api

import (
	"github.com/Ptt-official-app/go-openbbsmiddleware/apitypes"
	"github.com/Ptt-official-app/go-openbbsmiddleware/schema"
	"github.com/Ptt-official-app/go-openbbsmiddleware/types"
	"github.com/Ptt-official-app/go-openbbsmiddleware/utils"
	pttbbsapi "github.com/Ptt-official-app/go-pttbbs/api"
	"github.com/Ptt-official-app/go-pttbbs/bbs"
	"github.com/Ptt-official-app/go-pttbbs/ptttype"
	"github.com/gin-gonic/gin"
)

const LOAD_GENERAL_ARTICLES_R = "/board/:bid/articles"

type LoadGeneralArticlesParams struct {
	Keyword    string `json:"title,omitempty" form:"title,omitempty" url:"title,omitempty"`
	StartIdx   string `json:"start_idx,omitempty" form:"start_idx,omitempty" url:"start_idx,omitempty"`
	Max        int    `json:"limit,omitempty" form:"limit,omitempty" url:"limit,omitempty"`
	Descending bool   `json:"desc,omitempty"  form:"desc,omitempty" url:"desc,omitempty"`
}

type LoadGeneralArticlesPath struct {
	FBoardID apitypes.FBoardID `uri:"bid"`
}

type LoadGeneralArticlesResult struct {
	List           []*apitypes.ArticleSummary `json:"list"`
	NextIdx        string                     `json:"next_idx"`
	NextCreateTime types.Time8                `json:"next_create_time"`
	StartNumIdx    ptttype.SortIdx            `json:"start_num_idx"`
}

func NewLoadGeneralArticlesParams() *LoadGeneralArticlesParams {
	return &LoadGeneralArticlesParams{
		Max:        DEFAULT_MAX_LIST,
		Descending: DEFAULT_DESCENDING,
	}
}

func LoadGeneralArticlesWrapper(c *gin.Context) {
	params := NewLoadGeneralArticlesParams()
	path := &LoadGeneralArticlesPath{}
	LoginRequiredPathQuery(LoadGeneralArticles, params, path, c)
}

func LoadGeneralArticles(remoteAddr string, userID bbs.UUserID, params interface{}, path interface{}, c *gin.Context) (result interface{}, statusCode int, err error) {
	theParams, ok := params.(*LoadGeneralArticlesParams)
	if !ok {
		return nil, 400, ErrInvalidParams
	}

	thePath, ok := path.(*LoadGeneralArticlesPath)
	if !ok {
		return nil, 400, ErrInvalidPath
	}

	boardID, err := toBoardID(thePath.FBoardID, remoteAddr, userID, c)
	if err != nil {
		return nil, 500, err
	}

	//backend load-general-articles
	theParams_b := &pttbbsapi.LoadGeneralArticlesParams{
		StartIdx:  theParams.StartIdx,
		NArticles: theParams.Max,
		Desc:      theParams.Descending,
	}
	var result_b *pttbbsapi.LoadGeneralArticlesResult

	urlMap := map[string]string{
		"bid": string(boardID),
	}
	url := utils.MergeURL(urlMap, pttbbsapi.LOAD_GENERAL_ARTICLES_R)
	statusCode, err = utils.BackendGet(c, url, theParams_b, nil, &result_b)
	if err != nil || statusCode != 200 {
		return nil, statusCode, err
	}

	//update to db
	updateNanoTS := types.NowNanoTS()
	articleSummaries_db, userReadArticleMap, err := deserializeArticlesAndUpdateDB(userID, boardID, result_b.Articles, updateNanoTS)
	if err != nil {
		return nil, 500, err
	}

	userReadArticleMap, err = checkReadArticles(userID, boardID, userReadArticleMap, articleSummaries_db)
	if err != nil {
		return nil, 500, err
	}

	r := NewLoadGeneralArticlesResult(articleSummaries_db, userReadArticleMap, result_b)

	//update user_read_board
	if result_b.IsNewest {
		err = updateUserReadBoard(userID, boardID, updateNanoTS)
		if err != nil {
			return nil, 500, err
		}
	}

	return r, 200, nil
}

func checkReadArticles(userID bbs.UUserID, boardID bbs.BBoardID, userReadArticleMap map[bbs.ArticleID]bool, theList []*schema.ArticleSummary) (newUserReadArticleMap map[bbs.ArticleID]bool, err error) {
	queryArticleIDs := make([]bbs.ArticleID, 0, len(theList))
	checkArticleIDMap := make(map[bbs.ArticleID]int)
	for idx, each := range theList {
		isRead, ok := userReadArticleMap[each.ArticleID]
		if ok && isRead {
			continue
		}

		//check with read-time
		checkArticleIDMap[each.ArticleID] = idx
		queryArticleIDs = append(queryArticleIDs, each.ArticleID)
	}

	dbResults, err := schema.FindUserReadArticles(userID, boardID, queryArticleIDs)
	if err != nil {
		return nil, err
	}

	//setup read in the list
	//no need to update db, because we don't read the article yet.
	//the Read flag is set based on the existing db.UpdateNanoTS
	for _, each := range dbResults {
		eachArticleID := each.ArticleID
		eachReadNanoTS := each.UpdateNanoTS

		listIdx, ok := checkArticleIDMap[eachArticleID]
		if !ok {
			continue
		}

		eachInTheList := theList[listIdx]
		eachPostNanoTS := eachInTheList.CreateTime
		isRead := eachReadNanoTS > eachPostNanoTS
		userReadArticleMap[eachArticleID] = isRead
	}

	return userReadArticleMap, nil
}

func updateUserReadBoard(userID bbs.UUserID, boardID bbs.BBoardID, updateNanoTS types.NanoTS) (err error) {

	userReadBoard := &schema.UserReadBoard{UserID: userID, BBoardID: boardID, UpdateNanoTS: updateNanoTS}

	err = schema.UpdateUserReadBoard(userReadBoard)
	if err != nil {
		return err
	}

	return nil
}

func NewLoadGeneralArticlesResult(a_db []*schema.ArticleSummary, userReadArticleMap map[bbs.ArticleID]bool, result_b *pttbbsapi.LoadGeneralArticlesResult) *LoadGeneralArticlesResult {
	theList := make([]*apitypes.ArticleSummary, len(a_db))
	for i, each_db := range a_db {
		theList[i] = apitypes.NewArticleSummary(each_db)
		articleID := each_db.ArticleID
		isRead, ok := userReadArticleMap[articleID]
		if ok && isRead {
			theList[i].Read = true
		}
	}

	return &LoadGeneralArticlesResult{
		List:           theList,
		NextIdx:        result_b.NextIdx,
		NextCreateTime: types.Time8(result_b.NextCreateTime),
		StartNumIdx:    result_b.StartNumIdx,
	}
}
