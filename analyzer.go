package apiroper

import (
	"regexp"
	"strconv"
)

var r *regexp.Regexp
var ra *regexp.Regexp

func init() {
	r = regexp.MustCompile(`<<(?:([^\.\<\>]*)\.)?([^\<\>]+)>>`)
	ra = regexp.MustCompile(`[^\.\[\]]+`)
}

// analyze分析标记
func analyze(data string) (arguments []*argument) {
	arguments = []*argument{}
	matches := r.FindAllStringSubmatch(data, -1)
	argumentstr := ""
	for _, amatch := range matches {
		base := ""
		if len(amatch) >= 3 {
			base = amatch[1]
			argumentstr = amatch[2]
		} else if len(amatch) == 2 {
			argumentstr = amatch[1]
		}
		idkey := amatch[0]

		// 父参数指针
		var pargument *argument
		// 参数指针
		var argumento *argument
		if base != "" {
			// 所有非输入参数，以base参数为父参数
			pargument = &argument{
				idkey: idkey,
				base:  base,
				pkey:  base, // base参数只能以map-key标记
			}
		}
		args := ra.FindAllString(argumentstr, -1)
		for _, akey := range args {
			argumento = &argument{
				idkey:  idkey,
				base:   base,
				parent: pargument,
			}
			if pindex, err := strconv.Atoi(akey); err == nil {
				// pindex 是整形，父参数是slice
				if argumento.parent != nil {
					argumento.parent.argtype = ARGS_KEY_TYPE_SLICE
				}
				argumento.pindex = pindex
			} else {
				// 否则父参数是map
				if argumento.parent != nil {
					argumento.parent.argtype = ARGS_KEY_TYPE_MAP
				}
				argumento.pkey = akey
			}

			// 父指针指向当前
			pargument = argumento
		}
		arguments = append(arguments, argumento)
	}
	return
}

func findKeys(data string) []string {
	return r.FindAllString(data, -1)
}
