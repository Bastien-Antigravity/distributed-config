/*
ESSENTIAL PROCESS:
Thread-local error storage for the distributed-config CGO C ABI bridge.

DATA FLOW:
1. Input: Error code and C string message from CGO bridge functions.
2. Logic: Frees previous thread-local message and duplicates the new error string.
3. Output: Thread-local last_error and last_error_code state.

KEY PARAMETERS:
- code: Standardized DISTCONF_* error integer code.
- err: Null-terminated error description string.
*/

#include "helpers.h"

_Thread_local char* last_error = NULL;
_Thread_local int last_error_code = 0;

void set_last_error(int code, const char* err) {
    last_error_code = code;
    if (last_error != NULL) {
        free(last_error);
    }
    if (err == NULL) {
        last_error = NULL;
    } else {
        last_error = strdup(err);
    }
}
