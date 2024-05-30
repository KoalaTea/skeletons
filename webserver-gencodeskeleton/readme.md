# commands
ent and graphql are their own packages so they need to be made independently before other files can call on then. This is a little annoying so this provides a skeleton or requirements for that to work. 

createnv.sh will bootstrap create all the autogen code

# packages I find useful if starting from scratch again to get versioning right
entgo
gqlgen
otel
golang.org/x/oauth2
github.com/XSAM/otelsql
github.com/mattn/go-sqlite3