package model

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ErrNotFound = sqlx.ErrNotFound

// timeNow 返回当前时间，方便测试时 mock
var timeNow = time.Now
