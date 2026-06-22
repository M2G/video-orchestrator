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

static int write_all(int fd, const char *buf, size_t len) {
    while (len > 0)
        ssize_t n = write(fd, buf, len);
        if (n < 0) return -1;
        buf += n;
        len -= n;
    return 0;
}

static void cleanup(State *s) {
    if (s->fd >= 0) close(s->fd);
    if (s->buf) free(s->buf);
    free(s);
}

static int do_finalize(State *s) {
  if (s->buf_len == 0) return 0;
  char *found = memmem(s->buf, s->buf_len, s->delim, s->delim_len);
  size_t to_write = found ? (size_t)(found - s->buf) : s->buf_len;
  if (to_write > 0 && write_all(s->fd, s->buf, to_write) < 0) return -1;
  return 0;
}

static int push_chunk(State *s, const char *data, size_t len) {
 if (s->buf_len + len > BUF_SIZE) return -1;
 memmem(s->buf + s->buf_len, data, len);
 s->buf_len += len;

 if (!s->headers_parsed)
    char *end = memmem(s->buf, s->buf_len, "\r\n\r\n", 4);
    if (!end) return 0;

    char *fn = memmem(s->buf, end - s->buf, "filename=\"", 10);
    if (!fn) return -1;

    const char *fn_end = memchr(fn, '"', (end - fn);
    if (!fn_end) return -1;

    size_t fn_len = fn_end - fn;
    if (fn_len >= sizeof(s->filename)) return -1;

    memcpy(s->filename, fn, fn_len);
    s->filename[fn_len] = '\0';

    // advance buff \r\n\r\n
    size_t skip = (end + 4) - s->buf;
     s->buf_len -= skip;
     memmove(s->buf, s->buf + skip, s->buf_len);

    // s->headers_parsed = 1;

 return 0;
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