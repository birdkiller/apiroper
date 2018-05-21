package apiroper

// 合并map
func unionmap(maps ...map[string]*argument) map[string]*argument {
	newmap := map[string]*argument{}
	for _, m := range maps {
		for k, v := range m {
			newmap[k] = v
		}
	}
	return newmap
}
