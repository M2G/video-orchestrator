/*
Sources :

N-API :
https://nodejs.org/api/n-api.html : toute l'API C pour écrire des addons Node.js
https://nodejs.org/api/addons.html : guide général addons
https://github.com/nodejs/node-addon-examples
https://github.com/nodejs/node-addon-api

RFC Multipart :
https://www.rfc-editor.org/rfc/rfc7578 : spec multipart/form-data
https://www.rfc-editor.org/rfc/rfc2046 : spec multipart générale

Fonctions C (to use) :
memmem : https://man7.org/linux/man-pages/man3/memmem.3.html
memmove : https://en.cppreference.com/w/c/string/byte/memmove
open/write/close : https://man7.org/linux/man-pages/man2/open.2.html
*/

#include <node_api.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <fcntl.h>
#include <unistd.h>

#define BUFFER_SIZE (256 * 1024) // 256KB (chunks)

typedef struct {
  char     boundary[256];
  size_t   boundary_len;
  char     filepath[1024];
  int      fd;
  char    *buf;
  size_t   buf_len;
  size_t   buf_cap;
  int      headers_parsed;
  char     filename[256];
} ParseState;

static void cleanup(State *s) {
    if (s->fd >= 0) close(s->fd);
    if (s->buf) free(s->buf);
    free(s);
}


static void *find_mem(const void *haystack, size_t hlen,
                      const void *needle, size_t nlen) {

                        return NULL;
                      }

static int buf_append(ParseState *s, const char *data, size_t len) {
    return 0;
}

static int parse_headers(ParseState *s) {}

static int flush_safe(ParseState *s) {}

static napi_value CreateState(napi_env env, napi_callback_info info) {}

static napi_value PushChunk(napi_env env, napi_callback_info info) {}

static napi_value Finalize(napi_env env, napi_callback_info info) {}

static napi_value Init(napi_env env, napi_value exports) {}

NAPI_MODULE(NODE_GYP_MODULE_NAME, Init)