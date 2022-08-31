module github.com/je4/salon-digital/v2

go 1.18

replace github.com/je4/salon-digital/v2 => ./

replace github.com/je4/zsearch/v2 => ../zsearch/

//replace github.com/je4/PictureFS/v2 => ../PictureFS/

require (
	github.com/BurntSushi/toml v1.1.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/blevesearch/bleve/v2 v2.3.2
	github.com/chromedp/cdproto v0.0.0-20220515234810-83d799542a04
	github.com/chromedp/chromedp v0.8.1
	github.com/goph/emperror v0.17.2
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/je4/PictureFS/v2 v2.0.0-20220426160452-a0a705971ef3
	github.com/je4/bremote/v2 v2.0.0-20220513085231-c8464a3699b7
	github.com/je4/utils/v2 v2.0.6
	github.com/je4/zsearch/v2 v2.0.0-20220510114708-d9dcc9770f6c
	github.com/lmas/Damerau-Levenshtein v1.0.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	golang.org/x/text v0.3.7
)

require (
	cloud.google.com/go/compute v1.6.1 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20220517143526-88bb52951d5b // indirect
	github.com/RoaringBitmap/roaring v1.0.0 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/bits-and-blooms/bitset v1.2.2 // indirect
	github.com/blevesearch/bleve_index_api v1.0.1 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/gtreap v0.1.1 // indirect
	github.com/blevesearch/mmap-go v1.0.3 // indirect
	github.com/blevesearch/scorch_segment_api/v2 v2.1.0 // indirect
	github.com/blevesearch/segment v0.9.0 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/upsidedown_store_api v1.0.1 // indirect
	github.com/blevesearch/vellum v1.0.7 // indirect
	github.com/blevesearch/zapx/v11 v11.3.3 // indirect
	github.com/blevesearch/zapx/v12 v12.3.3 // indirect
	github.com/blevesearch/zapx/v13 v13.3.3 // indirect
	github.com/blevesearch/zapx/v14 v14.3.3 // indirect
	github.com/blevesearch/zapx/v15 v15.3.3 // indirect
	github.com/bluele/gcache v0.0.2 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/channelmeter/iso8601duration v0.0.0-20150204201828-8da3af7a2a61 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/dgraph-io/badger/v3 v3.2103.2 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/drgrib/maps v0.0.0-20220318162102-37b53c75ae89 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/go-git/go-git/v5 v5.4.2 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v2.0.6+incompatible // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/sessions v1.2.1 // indirect
	github.com/htfy96/reformism v0.0.0-20160819020323-e5bfca398e73 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/je4/sitemap v1.0.1-0.20210914120028-a4ef87562716 // indirect
	github.com/je4/zsync v0.0.0-20211108172845-6b701afd5ef8 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/juliangruber/go-intersect v1.1.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.15.4 // indirect
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	github.com/lib/pq v1.10.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/minio-go/v7 v7.0.26 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/opensearch-project/opensearch-go v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/xid v1.4.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/snabb/diagio v1.0.0 // indirect
	github.com/vanng822/go-solr v0.10.0 // indirect
	github.com/xanzy/ssh-agent v0.3.1 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.mongodb.org/mongo-driver v1.9.1 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.0.0-20220518034528-6f7dac969898 // indirect
	golang.org/x/image v0.0.0-20220413100746-70e8d0d3baa9 // indirect
	golang.org/x/net v0.0.0-20220517181318-183a9ca12b87 // indirect
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	golang.org/x/sync v0.0.0-20220513210516-0976fa681c29 // indirect
	golang.org/x/sys v0.0.0-20220517195934-5e4e11fc645e // indirect
	google.golang.org/api v0.80.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220518221133-4f43b3371335 // indirect
	google.golang.org/grpc v1.46.2 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/resty.v1 v1.12.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)
