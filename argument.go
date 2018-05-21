package apiroper

import (
	"fmt"
)

const (
	ARGS_KEY_TYPE_NORMAL = iota
	ARGS_KEY_TYPE_SLICE
	ARGS_KEY_TYPE_MAP
)

type argument struct {
	argtype int
	idkey   string // 唯一标识(若标识相同则取值路径也相同)
	parent  *argument
	pkey    string // 从父参数中获取该参数所需的key(map)
	pindex  int    // 从父参数中获取该参数所需的下表(slice)
	base    string // 根节点
}

func (self *argument) getValue(source interface{}) (interface{}, error) {
	if self.parent != nil {
		pValue, err := self.parent.getValue(source)
		if err != nil {
			return nil, err
		}
		// 根据父节点的类型进行强转
		if self.parent.argtype == ARGS_KEY_TYPE_SLICE {
			// slice
			sslice, ok := pValue.([]interface{})
			if ok != true {
				return nil, fmt.Errorf("error when argument transfer (SLICE)")
			}
			return sslice[self.pindex], nil
		} else if self.parent.argtype == ARGS_KEY_TYPE_MAP {
			// map
			smap, ok := pValue.(map[string]interface{})
			if ok != true {
				return nil, fmt.Errorf("error when argument transfer (MAP)")
			}
			return smap[self.pkey], nil
		} else {
			// 无类型，直接返回父值
			return pValue, nil
		}
	} else {
		// base节点，肯定是map-key形式
		smap, ok := source.(map[string]interface{})
		if ok != true {
			return nil, fmt.Errorf("error when argument transfer (BASE)")
		}
		return smap[self.pkey], nil
	}
}
