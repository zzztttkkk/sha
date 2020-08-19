# suna
a web framework based on [fasthttp](https://github.com/valyala/fasthttp).

## example
[example](https://github.com/zzztttkkk/suna/tree/master/example)

## sub modules
### auth
get user information from `fasthttp.RequestCtx`

### cache
- Lru: concurrent safe lru cache.
- RedCache: cache the entire response in Redis.

### config
provide a config type.

### middleware
- access logging
- cors
- rate limiter

### output
send json response and handler errors.

### rbac
a Role-based access control implementation.

- a role can inherit permissions from other roles. 
- no mutex between roles.
- no mutex between permissions.
- subject cannot have private permissions.

### readlock
a redis lock.

### reflectx
type reflect tools.

### secret
- some hash functions
- generate random bytes
- id token dump/load
- ase encode/decode

### session
redis session implementation.(if the authentication is passed, it will
try to use the previous session)

- captcha generate(image&audio)/verify, based on [captcha](https://github.com/dchest/captcha).

### sqls
some [sqlx](https://github.com/jmoiron/sqlx) wrapper functions, make execute sql easy.

### sqls.builder
some [sqrl](https://github.com/zzztttkkk/sqrl) wrapper functions, make generate sql easy.

### utils.toml
load toml files and do replace in string field.

- `file://***` replace with the file content.
- `$ENV{***}` replace with the environment variable.

### validator
- request form validate.
- generate markdown table as document.
