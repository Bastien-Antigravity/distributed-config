package core

// =============================================================================
// ESSENTIAL PROCESS:
// Recursive dictionary merge engine performing deep overrides of nested
// capability maps, file configurations, and runtime overrides.
//
// DATA FLOW:
// 1. Input: target and source map[string]interface{} structures.
// 2. Logic: Traverses nested maps recursively; source keys overwrite target keys.
// 3. Output: Merged map[string]interface{} containing the unified dictionary.
//
// KEY PARAMETERS:
// - target: Base destination map.
// - source: Overriding higher-precedence map.
// =============================================================================

// -----------------------------------------------------------------------------

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
