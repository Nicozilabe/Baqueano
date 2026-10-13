package drive

// Default OAuth client, compiled into baq so that `baq drive login` works
// without prompting the user for their own Google client.
//
// The values are NOT stored in the repository. They are injected at build time
// via -ldflags:
//
//   - CI / releases: GoReleaser reads the GitHub secrets BAQUEANO_GOOGLE_CLIENT_ID
//     and BAQUEANO_GOOGLE_CLIENT_SECRET (see .goreleaser.yaml and release.yml).
//   - Local builds: run `make build`, scripts/build.ps1 or scripts/build.sh,
//     which read the gitignored .env file (copy .env.example and fill it in).
//
// Desktop-app client secrets are not treated as confidential by Google: they
// ship inside every installed binary, which is why injecting them at build
// time is expected.
var (
	builtinClientID     = ""
	builtinClientSecret = ""
)
