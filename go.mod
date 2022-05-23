module github.com/paketo-buildpacks/jammy-tiny-stack

go 1.18

require (
	github.com/google/go-containerregistry v0.8.1-0.20220209165246-a44adc326839
	github.com/onsi/gomega v1.19.0
	github.com/paketo-buildpacks/jam v1.3.0
	github.com/paketo-buildpacks/occam v0.8.0
	github.com/paketo-buildpacks/packit v1.3.1
	github.com/paketo-buildpacks/packit/v2 v2.3.0
	github.com/sclevine/spec v1.4.0
)

require (
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/ForestEckhardt/freezer v0.0.11 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.11.4 // indirect
	github.com/gabriel-vasile/mimetype v1.4.0 // indirect
	github.com/klauspost/compress v1.15.2 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20220114050600-8b9d41f48198 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/vbatts/tar-split v0.11.2 // indirect
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/paketo-buildpacks/occam v0.8.0 => github.com/paketo-buildpacks/occam v0.8.1-0.20220519174407-e2e5be747308
