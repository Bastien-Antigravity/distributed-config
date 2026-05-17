#ifndef HELPERS_H
#define HELPERS_H

#include <stdlib.h>
#include <string.h>

// Standardized Error Codes
#define DISTCONF_SUCCESS                0
#define DISTCONF_ERR_GENERIC            1
#define DISTCONF_ERR_INVALID_HANDLE      2
#define DISTCONF_ERR_KEY_NOT_FOUND       3
#define DISTCONF_ERR_VALIDATION_FAILED   4
#define DISTCONF_ERR_NETWORK_FAILURE     5
#define DISTCONF_ERR_DECRYPTION_FAILED   6
#define DISTCONF_ERR_INVALID_INPUT       7

extern char* last_error;
extern int last_error_code;

void set_last_error(int code, const char* err);

#endif
