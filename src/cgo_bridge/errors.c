#include "helpers.h"

char* last_error = NULL;

void set_last_error(const char* err) {
    if (last_error != NULL) {
        free(last_error);
    }
    if (err == NULL) {
        last_error = NULL;
    } else {
        last_error = strdup(err);
    }
}
