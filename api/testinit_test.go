package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"

	"github.com/Ptt-official-app/go-openbbsmiddleware/db"
	"github.com/Ptt-official-app/go-openbbsmiddleware/queue"
	"github.com/Ptt-official-app/go-openbbsmiddleware/schema"
	"github.com/Ptt-official-app/go-openbbsmiddleware/types"
	"github.com/Ptt-official-app/go-openbbsmiddleware/utils"
	"github.com/Ptt-official-app/go-pttbbs/bbs"
	"github.com/gin-gonic/gin"
	"github.com/google/go-querystring/query"
	"github.com/sirupsen/logrus"
)

var (
	testIP = "127.0.0.1"
)

func setupTest() {
	utils.SetIsTest()

	db.SetIsTest()

	types.SetIsTest("api")

	schema.SetIsTest()

	queue.SetIsTest()

	SetIsTest()

	initTest()

	params := &RegisterClientParams{ClientID: "default_client_id", ClientType: types.CLIENT_TYPE_APP}
	_, _, _ = RegisterClient("localhost", bbs.UUserID("SYSOP"), params, nil)
	//logrus.Infof("api.setupTest: after RegisterClient: status: %v e: %v", statusCode, err)
}

func teardownTest() {
	UnsetIsTest()

	queue.UnsetIsTest()

	schema.UnsetIsTest()

	types.UnsetIsTest("api")

	db.UnsetIsTest()

	utils.UnsetIsTest()
}

func testSetRequest(reqPath string, pathPattern string, params interface{}, jwt string, csrfToken string, headers map[string]string, method string, f gin.HandlerFunc) (*httptest.ResponseRecorder, *gin.Context, *gin.Engine) {
	var jsonBytes []byte

	if method == "GET" {
		v, _ := query.Values(params)
		reqPath = reqPath + "?" + v.Encode()
	} else {
		jsonBytes, _ = json.Marshal(params)
	}

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	switch method {
	case "POST":
		r.POST(pathPattern, f)
	case "GET":
		r.GET(pathPattern, f)
	}

	req := httptest.NewRequest(method, reqPath, bytes.NewBuffer(jsonBytes))

	logrus.Infof("testSetRequest: method: %v reqPath: %v pathPattern: %v f: %v", method, reqPath, pathPattern, f)

	req.Header.Set("Host", "localhost:5678")
	if jwt != "" {
		req.Header.Set("Authorization", "bearer "+jwt)
	}

	if csrfToken != "" {
		req.Header.Set("X-CSRFToken", csrfToken)
	}

	for key, val := range headers {
		req.Header.Set(key, val)
	}

	c.Request = req

	return w, c, r
}
