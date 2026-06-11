package core

// DeepMerge recursively merges source into target.
// Entries in source override entries in target.
func DeepMerge(target, source map[string]interface{}) map[string]interface{} {
	if target == nil {
		target = make(map[string]interface{})
	}
	for k, v := range source {
		if sourceMap, ok := v.(map[string]interface{}); ok {
			if targetMap, ok := target[k].(map[string]interface{}); ok {
				target[k] = DeepMerge(targetMap, sourceMap)
			} else {
				target[k] = v
			}
		} else {
			target[k] = v
		}
	}
	return target
}
