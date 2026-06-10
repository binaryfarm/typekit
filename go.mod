module github.com/binaryfarm/typekit

go 1.26

require (
	github.com/dlclark/regexp2 v1.11.4
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible
	github.com/google/pprof v0.0.0-20230207041349-798e818bf904
	golang.org/x/text v0.38.0
)

require github.com/microsoft/typescript-go v0.0.0-20260609232358-f20e9196e66c // indirect

require (
	github.com/evanw/esbuild v0.25.8
	golang.org/x/sys v0.46.0 // indirect
)

replace github.com/microsoft/typescript-go => ./vendor-ts
