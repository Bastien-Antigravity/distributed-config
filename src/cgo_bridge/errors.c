#include "helpers.h"

char* last_error = NULL;
int last_error_code = 0;

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
