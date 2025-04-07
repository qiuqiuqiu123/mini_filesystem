package client

import "mini_filesystem/common"

type Err struct {
	idx int
	loc *common.Location
	err error
}
