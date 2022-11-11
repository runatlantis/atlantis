module github.com/runatlantis/atlantis

go 1.19

replace google.golang.org/grpc => google.golang.org/grpc v1.45.0

require (
	cloud.google.com/go v0.81.0 // indirect
	cloud.google.com/go/storage v1.10.0 // indirect
	github.com/Laisky/graphql v1.0.5
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v12 v12.0.0 // indirect
	github.com/aws/aws-sdk-go v1.40.12
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/bradleyfalzon/ghinstallation v1.1.1
	github.com/briandowns/spinner v0.0.0-20170614154858-48dbb65d7bd5
	github.com/cpuguy83/go-md2man/v2 v2.0.0-20190314233015-f79a8a8ca69d // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v0.0.0-20180620051407-e2593239d949
	github.com/elazarl/go-bindata-assetfs v1.0.1
	github.com/fatih/color v1.7.0 // indirect
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568
	github.com/fsnotify/fsnotify v1.4.10-0.20200417215612-7f4cf4dd2b52 // indirect
	github.com/go-ozzo/ozzo-validation v0.0.0-20170913164239-85dcd8368eba
	github.com/go-playground/locales v0.12.1 // indirect
	github.com/go-playground/universal-translator v0.16.0 // indirect
	github.com/go-test/deep v1.0.7
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/go-github/v29 v29.0.2 // indirect
	github.com/google/go-github/v45 v45.2.0
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-getter v1.5.7
	github.com/hashicorp/go-retryablehttp v0.6.8 // indirect
	github.com/hashicorp/go-safetemp v1.0.0 // indirect
	github.com/hashicorp/go-version v1.5.0
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/hcl/v2 v2.6.0 // indirect
	github.com/hashicorp/terraform-config-inspect v0.0.0-20200806211835-c481b8bfa41e
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/klauspost/compress v1.11.2 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/lusis/slack-test v0.0.0-20190426140909-c40012f20018 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mcdafydd/go-azuredevops v0.12.0
	github.com/microcosm-cc/bluemonday v1.0.15
	github.com/mitchellh/colorstring v0.0.0-20150917214807-8631ce90f286
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170603005431-491d3605edfb
	github.com/nlopes/slack v0.4.0
	github.com/onsi/gomega v1.10.1 // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/petergtz/pegomock v2.9.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/remeh/sizedwaitgroup v1.0.0
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/shurcooL/githubv4 v0.0.0-20220520033151-0b4e3294ff00
	github.com/shurcooL/graphql v0.0.0-20181231061246-d48a9a75455f // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.6.1-0.20200528085638-6699a89a232f // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.0-20170905172051-b78744579491
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.8.1
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/thomaspoignant/go-feature-flag v0.18.4
	github.com/ulikunitz/xz v0.5.8 // indirect
	github.com/urfave/cli v1.22.5
	github.com/urfave/negroni v0.3.0
	github.com/xanzy/go-gitlab v0.50.3
	github.com/zclconf/go-cty v1.10.0 // indirect
	go.etcd.io/bbolt v1.3.6
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.6.0 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602 // indirect
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/tools v0.2.1-0.20221101170700-b5bc717366b2 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/api v0.44.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220602131408-e326c6e8e9c8 // indirect
	google.golang.org/grpc v1.47.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools v2.2.0+incompatible // indirect
)

require (
	github.com/antlr/antlr4 v0.0.0-20201206235148-c87e55b61113 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/cactus/go-statsd-client/statsd v0.0.0-20200623234511-94959e3146b2
	github.com/nikunjy/rules v0.0.0-20200120082459-0b7c4dc9dc86 // indirect
	github.com/twmb/murmur3 v1.1.6 // indirect
)

require go.temporal.io/sdk v1.15.0

require (
	github.com/aws/aws-sdk-go-v2 v1.13.0
	github.com/aws/aws-sdk-go-v2/config v1.13.1
	github.com/aws/aws-sdk-go-v2/service/sqs v1.16.0
	github.com/graymeta/stow v0.2.7
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/uber-go/tally/v4 v4.1.2
	go.temporal.io/sdk/contrib/tally v0.1.0
	logur.dev/adapter/zap v0.5.0
	logur.dev/logur v0.17.0
)

require (
	github.com/BurntSushi/toml v1.2.0 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/bradleyfalzon/ghinstallation/v2 v2.1.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rs/zerolog v1.27.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/exp/typeparams v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/tools/gopls v0.10.1 // indirect
	golang.org/x/vuln v0.0.0-20221010193109-563322be2ea9 // indirect
	honnef.co/go/tools v0.3.3 // indirect
	mvdan.cc/gofumpt v0.3.1 // indirect
	mvdan.cc/xurls/v2 v2.4.0 // indirect
)

require (
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.10.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.14.0 // indirect
	github.com/aws/smithy-go v1.10.0 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gogo/status v1.1.0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/hashicorp/terraform-json v0.14.0
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/palantir/go-githubapp v0.13.1
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	go.temporal.io/api v1.8.0 // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
	golang.org/x/term v0.1.0
)
