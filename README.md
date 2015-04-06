#KVB
Key-Value Blog

Go, HTML Templates, BoltDB

##Recommendations (in order):
* GVM https://github.com/moovweb/gvm
* Pathogen https://github.com/tpope/vim-pathogen
* vim-go https://github.com/fatih/vim-go
##Requirements:
* Installation of Go and environment variables (https://golang.org/doc/install if you don't want to use gvm)
* BoltDB https://github.com/boltdb/bolt `go get github.com/boltdb/bolt/...`

##Description:
X - section - highest level of data (ex. about, posts, contact)
Y - post - specific blog posts
b - browse - display the section or post in a section
e - edit - open a section for editing

Examples:
kvb.com/b/X
kvb.com/b/X/Y
kvb.com/e/X/
kvb.com/e/X/Y
...

"b" and "e" are directly related to html templates that allow users to browse and edit respectively

X directly correlates to a bucket in boltdb
Y directly correlates to a title and key(body) in boltdb
