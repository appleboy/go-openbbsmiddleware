package api

import (
	"bytes"
	"strconv"

	"github.com/Ptt-official-app/go-openbbsmiddleware/apitypes"
	"github.com/Ptt-official-app/go-openbbsmiddleware/fav"
	"github.com/Ptt-official-app/go-openbbsmiddleware/schema"
	"github.com/Ptt-official-app/go-openbbsmiddleware/types"
	"github.com/Ptt-official-app/go-openbbsmiddleware/utils"
	pttbbsapi "github.com/Ptt-official-app/go-pttbbs/api"
	"github.com/Ptt-official-app/go-pttbbs/bbs"
	pttbbsfav "github.com/Ptt-official-app/go-pttbbs/ptt/fav"
	"github.com/Ptt-official-app/go-pttbbs/ptttype"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func tryGetUserFavorites(
	userID bbs.UUserID,
	levelIdx schema.LevelIdx,
	startIdxStr string,
	ascending bool,
	limit int,
	c *gin.Context) (

	userFavorites []*schema.UserFavorites,
	nextIdxStr string,
	statusCode int,
	err error) {

	startIdx := 0
	if len(startIdxStr) > 0 {
		startIdx, err = strconv.Atoi(startIdxStr)
		if err != nil {
			return nil, "", 400, ErrInvalidParams
		}
	}

	//user-favorites-meta
	userFavoritesMeta, err := schema.GetUserFavoritesMeta(userID)
	if err != nil {
		return nil, "", 500, err
	}

	dbMTime := types.NanoTS(0)
	newDoubleBufferIdx := 0
	if userFavoritesMeta != nil {
		dbMTime = userFavoritesMeta.MTime
		newDoubleBufferIdx = (userFavoritesMeta.DoubleBufferIdx + 1) % schema.N_USER_FAVORITES_DOUBLE_BUFFER
	}

	//backend get user-favorites
	theParams_b := &pttbbsapi.GetFavoritesParams{
		RetrieveTS: dbMTime.ToTime4(),
	}
	var result_b *pttbbsapi.GetFavoritesResult

	urlMap := map[string]string{
		"uid": string(userID),
	}
	url := utils.MergeURL(urlMap, pttbbsapi.GET_FAV_R)

	statusCode, err = utils.BackendGet(c, url, theParams_b, nil, &result_b)
	if err != nil {
		return nil, "", statusCode, err
	}

	//check mtime
	updateNanoTS := types.NowNanoTS()

	backendMTime := types.Time4ToNanoTS(result_b.MTime)
	if backendMTime > dbMTime {
		err = deserializeUserFavoritesAndUpdateDB(userID, backendMTime, result_b.Content, updateNanoTS, newDoubleBufferIdx)
		if err != nil {
			return nil, "", 500, err
		}

	}

	//get db
	userFavorites, nextIdx, err := getUserFavoritesFromDB(userID, levelIdx, startIdx, ascending, limit)
	if err != nil {
		return nil, "", 500, err
	}

	nextIdxStr = ""
	if nextIdx >= 0 {
		nextIdxStr = strconv.Itoa(nextIdx)
	}

	return userFavorites, nextIdxStr, 200, nil
}

func deserializeUserFavoritesAndUpdateDB(userID bbs.UUserID, mTime types.NanoTS, contentWithVersion []byte, updateNanoTS types.NanoTS, doubleBufferIdx int) (err error) {
	if len(contentWithVersion) < 2 {
		return ErrInvalidFav
	}
	content := contentWithVersion[2:]

	file := bytes.NewReader(content)

	f, err := fav.ReadFavrec(file, nil, nil, 0)
	if err != nil {
		return err
	}

	meta, userFavorites := schema.FavToUserFavorites(f, userID, doubleBufferIdx, updateNanoTS, mTime)

	err = schema.UpdateUserFavorites(userID, doubleBufferIdx, userFavorites, mTime, updateNanoTS)
	if err != nil {
		return err
	}

	//ok if data is out-dated.
	err = schema.UpdateUserFavoritesMeta(meta)
	if err == schema.ErrNoMatch {
		err = nil
	}
	if err != nil {
		return err
	}

	return nil
}

func getUserFavoritesFromDB(userID bbs.UUserID, levelIdx schema.LevelIdx, startIdx int, ascending bool, limit int) (userFavorites []*schema.UserFavorites, nextIdx int, err error) {

	metaSumamry, err := schema.GetUserFavoritesMetaSummary(userID)
	if err == mongo.ErrNoDocuments {
		return nil, -1, nil
	}
	if err != nil {
		return nil, 0, err
	}

	userFavorites, err = schema.GetUserFavorites(userID, metaSumamry.DoubleBufferIdx, levelIdx, startIdx, ascending, limit+1, metaSumamry.MTime)
	if err != nil {
		return nil, 0, err
	}
	if len(userFavorites) <= limit {
		nextIdx = -1
	} else {
		nextIdx = userFavorites[limit].Idx
		userFavorites = userFavorites[:limit]
	}

	return userFavorites, nextIdx, nil
}

func tryGetBoardSummaryMapFromUserFavorites(userID bbs.UUserID, userFavorites_db []*schema.UserFavorites, c *gin.Context) (boardSummaryMap_db map[int]*schema.BoardSummary, userBoardInfoMap map[bbs.BBoardID]*apitypes.UserBoardInfo, statusCode int, err error) {
	bids := bidsInUserFavorites(userFavorites_db)

	//backend get boards by bids
	theParams_b := &pttbbsapi.LoadBoardsByBidsParams{
		Bids: bids,
	}
	var result_b *pttbbsapi.LoadBoardsByBidsResult

	url := pttbbsapi.LOAD_BOARDS_BY_BIDS_R
	statusCode, err = utils.BackendPost(c, url, theParams_b, nil, &result_b)
	if err != nil || statusCode != 200 {
		return nil, nil, statusCode, err
	}

	//update to db
	updateNanoTS := types.NowNanoTS()
	boardSummaries_db, userBoardInfoMap, err := deserializeBoardsAndUpdateDB(userID, result_b.Boards, updateNanoTS)
	if err != nil {
		return nil, nil, 500, err
	}

	boardSummaryMap_db = map[int]*schema.BoardSummary{}
	for _, each := range boardSummaries_db {
		theID, _, err := each.BBoardID.ToRaw()
		if err != nil {
			continue
		}
		boardSummaryMap_db[int(theID)] = each
	}

	return boardSummaryMap_db, userBoardInfoMap, 200, nil
}

func bidsInUserFavorites(userFavorites []*schema.UserFavorites) (bids []ptttype.Bid) {
	bids = make([]ptttype.Bid, 0, len(userFavorites))
	for _, each := range userFavorites {
		if each.TheType != pttbbsfav.FAVT_BOARD {
			continue
		}

		eachBid := ptttype.Bid(each.TheID)

		bids = append(bids, eachBid)
	}

	return bids
}
