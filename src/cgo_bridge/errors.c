/*
ESSENTIAL PROCESS:
Thread-local and fallback fixed-buffer error storage for the distributed-config CGO C ABI bridge.

DATA FLOW:
1. Input: Error code and C string message from CGO bridge functions.
2. Logic: Copies message into fixed-size thread-local and mutex-guarded global buffers via snprintf (zero strdup leaks).
3. Output: Safe const char* error descriptions and integer error codes across Go scheduler thread switches.

KEY PARAMETERS:
- code: Standardized DISTCONF_* error integer code.
- err: Null-terminated error description string.
*/

#include "helpers.h"
#include <pthread.h>
#include <stdio.h>

#define DISTCONF_ERR_BUF_SIZE 512

static _Thread_local char last_error_buf[DISTCONF_ERR_BUF_SIZE] = {0};
static _Thread_local int last_error_code = 0;

static pthread_mutex_t global_err_mu = PTHREAD_MUTEX_INITIALIZER;
static char global_last_error_buf[DISTCONF_ERR_BUF_SIZE] = {0};
static int global_last_error_code = 0;

// -----------------------------------------------------------------------------

void set_last_error(int code, const char* err) {
    last_error_code = code;
    if (err == NULL || err[0] == '\0') {
        last_error_buf[0] = '\0';
    } else {
        snprintf(last_error_buf, sizeof(last_error_buf), "%s", err);
    }

    pthread_mutex_lock(&global_err_mu);
    global_last_error_code = code;
    if (err == NULL || err[0] == '\0') {
        global_last_error_buf[0] = '\0';
    } else {
        snprintf(global_last_error_buf, sizeof(global_last_error_buf), "%s", err);
    }
    pthread_mutex_unlock(&global_err_mu);
}

// -----------------------------------------------------------------------------

const char* get_last_error(void) {
    if (last_error_code != 0 && last_error_buf[0] != '\0') {
        return last_error_buf;
    }
    pthread_mutex_lock(&global_err_mu);
    const char* res = (global_last_error_code != 0 && global_last_error_buf[0] != '\0') ? global_last_error_buf : NULL;
    pthread_mutex_unlock(&global_err_mu);
    return res;
}

// -----------------------------------------------------------------------------

int get_last_error_code(void) {
    if (last_error_code != 0) {
        return last_error_code;
    }
    pthread_mutex_lock(&global_err_mu);
    int code = global_last_error_code;
    pthread_mutex_unlock(&global_err_mu);
    return code;
}
