package drive

// Default OAuth client, compiled into sshm so that every build works out of
// the box without the user creating their own Google client - including a
// plain local `go build` / `go run .`.
//
// Fill in the Client ID and Client Secret of a Google Cloud OAuth client of
// type "Desktop app" (https://console.cloud.google.com/apis/credentials).
//
// A desktop-app client secret is NOT treated as confidential by Google: it is
// expected to ship inside every installed binary, so committing it here is
// fine. Official releases may still override these values at build time with
//
//	-ldflags "-X github.com/Gu1llaum-3/sshm/internal/drive.builtinClientID=..."
//
// (see .goreleaser.yaml), which wins over the literals below.
var (
	builtinClientID     = ""
	builtinClientSecret = ""
)
