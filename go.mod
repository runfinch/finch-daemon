module github.com/runfinch/finch-daemon

go 1.24.11

require (
	github.com/containerd/cgroups/v3 v3.1.2
	github.com/containerd/containerd/api v1.10.0
	github.com/containerd/containerd/v2 v2.2.1
	github.com/containerd/errdefs v1.0.0
	github.com/containerd/fifo v1.1.0
	github.com/containerd/go-cni v1.1.13
	github.com/containerd/log v0.1.0
	github.com/containerd/nerdctl/v2 v2.2.1
	github.com/containerd/platforms v1.0.0-rc.2
	github.com/containerd/typeurl/v2 v2.2.3
	github.com/containernetworking/cni v1.3.0
	github.com/coreos/go-iptables v0.8.0
	github.com/coreos/go-systemd/v22 v22.7.0
	github.com/distribution/reference v0.6.0
	github.com/docker/cli v29.1.3+incompatible
	github.com/docker/docker v28.5.2+incompatible
	github.com/docker/go-connections v0.6.0
	github.com/docker/go-units v0.5.0
	github.com/getlantern/httptest v0.0.0-20161025015934-4b40f4c7e590
	github.com/gofrs/flock v0.13.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/moby/go-archive v0.2.0
	github.com/moby/moby v28.5.2+incompatible
	github.com/moby/sys/user v0.4.0
	github.com/onsi/ginkgo/v2 v2.27.5
	github.com/onsi/gomega v1.39.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.1
	github.com/opencontainers/runtime-spec v1.3.0
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/pkg/errors v0.9.1
	github.com/runfinch/common-tests v0.10.4
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/sirupsen/logrus v1.9.4
	github.com/spf13/afero v1.15.0
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.11.1
	github.com/vishvananda/netlink v1.3.1
	github.com/vishvananda/netns v0.0.5
	go.uber.org/mock v0.6.0
	golang.org/x/net v0.49.0
	golang.org/x/sys v0.40.0
	google.golang.org/protobuf v1.36.11
)

require (
	cyphar.com/go-pathrs v0.2.2 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/nerdctl/mod/tigron v0.0.0-20250720235051-d775a8c42fbb // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/dsig v1.0.0 // indirect
	github.com/lestrrat-go/dsig-secp256k1 v1.0.0 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc/v3 v3.0.2 // indirect
	github.com/lestrrat-go/jwx/v3 v3.0.12 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20251013123823-9fd1530e3ec3 // indirect
	github.com/moby/sys/capability v0.4.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.4 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shoenig/go-m1cpu v0.1.7 // indirect
	github.com/tchap/go-patricia/v2 v2.3.3 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/valyala/fastjson v1.6.7 // indirect
	github.com/vektah/gqlparser/v2 v2.5.31 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yashtewari/glob-intersection v0.2.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/sdk v1.39.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251222181119-0a764e51fe1b // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.14.0-rc.1 // indirect
	github.com/cilium/ebpf v0.20.0 // indirect
	github.com/containerd/accelerated-container-image v1.3.0 // indirect
	github.com/containerd/console v1.0.5 // indirect
	github.com/containerd/continuity v0.4.5 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/go-runc v1.1.0 // indirect
	github.com/containerd/imgcrypt/v2 v2.0.2 // indirect
	github.com/containerd/nydus-snapshotter v0.15.10 // indirect
	github.com/containerd/plugin v1.0.0 // indirect
	github.com/containerd/stargz-snapshotter v0.18.1 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.18.1 // indirect
	github.com/containerd/stargz-snapshotter/ipfs v0.18.1 // indirect
	github.com/containerd/ttrpc v1.2.7 // indirect
	github.com/containernetworking/plugins v1.9.0 // indirect
	github.com/containers/ocicrypt v1.2.1 // indirect
	github.com/cyphar/filepath-securejoin v0.6.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/djherbis/times v1.6.0 // indirect
	github.com/docker/docker-credential-helpers v0.9.5
	github.com/fahedouch/go-logrotate v0.3.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fluent/fluent-logger-golang v1.10.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/getlantern/mockconn v0.0.0-20200818071412-cb30d065a848 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/godbus/dbus/v5 v5.2.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20251213031049-b05bdaca462f // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/ipfs/go-cid v0.6.0 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	github.com/moby/sys/mount v0.3.4 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/signal v0.7.1 // indirect
	github.com/moby/sys/symlink v0.3.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.1.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.16.1 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-varint v0.1.0 // indirect
	github.com/open-policy-agent/opa v1.12.3
	github.com/opencontainers/runtime-tools v0.9.1-0.20251205004911-5e639034dcdc // indirect
	github.com/opencontainers/selinux v1.13.1 // indirect
	github.com/petermattis/goid v0.0.0-20251121121749-a11dd1a45f9a // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rootless-containers/bypass4netns v0.4.2 // indirect
	github.com/rootless-containers/rootlesskit/v2 v2.3.6 // indirect
	github.com/sasha-s/go-deadlock v0.3.6 // indirect
	github.com/smallstep/pkcs7 v0.2.1 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stefanberger/go-pkcs11uri v0.0.0-20230803200340-78284954bff6 // indirect
	github.com/tinylib/msgp v1.6.1 // indirect
	github.com/vbatts/tar-split v0.12.2 // indirect
	github.com/yuchanns/srslog v1.1.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.64.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/exp v0.0.0-20251219203646-944ab1f22d93 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/term v0.39.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251222181119-0a764e51fe1b // indirect
	google.golang.org/grpc v1.77.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.4.1 // indirect
	tags.cncf.io/container-device-interface v1.1.0 // indirect
	tags.cncf.io/container-device-interface/specs-go v1.1.0 // indirect
)

exclude github.com/containerd/nerdctl/mod/tigron v0.0.0

replace cyphar.com/go-pathrs => github.com/cyphar/libpathrs/go-pathrs v0.2.1

// Pin docker cli version to v28.5.2+incompatible until we migrate to moby/moby/v2
replace github.com/docker/cli => github.com/docker/cli v28.5.2+incompatible
