# lazyspotify

`lazyspotify` is a terminal Spotify client that uses a patched `go-librespot`
daemon for playback.

## Runtime requirements

- A Spotify Premium account.
- A patched `go-librespot` daemon build from
  `https://github.com/dubeyKartikay/go-librespot/`.
- A working system keyring.
  On Linux this is a hard requirement: `lazyspotify` will not fall back to
  plaintext token storage if the keyring is unavailable.
- Linux clipboard integration is optional.
  For the auth screen's copy shortcut, install one of `wl-clipboard`, `xclip`,
  or `xsel`.

Package defaults:

- Linux package builds default the daemon audio backend to `alsa`.
- macOS package builds default the daemon audio backend to `audio-toolbox`.

## Daemon discovery

At startup, `lazyspotify` resolves the playback daemon in this order:

1. `librespot.daemon.cmd` from `config.yaml`
2. A packaged default daemon path compiled into the binary at build time

Not supported:

- `LAZYSPOTIFY_LIBRESPOT_DAEMON`
- `PATH` lookup
- probing next to the executable
- relocatable bundled daemon discovery

If you install `lazyspotify` without packaging and do not compile in a default
daemon path, you must set `librespot.daemon.cmd`.

## Installation

### Primary channels

- macOS: Homebrew tap formula.
- Ubuntu: PPA.
- Fedora: COPR.
- Arch: AUR `lazyspotify-bin`.

Fallback release assets are published on GitHub Releases:

- macOS: signed and notarized `.zip` archives.
- Ubuntu: `.deb`.
- Fedora: `.rpm`.
- Arch: binary tarball consumed by `lazyspotify-bin`.

### Packaged install layout

Every package ships both binaries:

- `lazyspotify`
- `lazyspotify-librespot`

Compile the daemon path into the `lazyspotify` binary with
`github.com/dubeyKartikay/lazyspotify/buildinfo.PackagedDaemonPath`.

Install layouts:

- Homebrew: `bin/lazyspotify` and `#{opt_libexec}/lazyspotify-librespot`
- Ubuntu: `/usr/bin/lazyspotify` and `/usr/lib/lazyspotify/lazyspotify-librespot`
- Fedora: `/usr/bin/lazyspotify` and `/usr/libexec/lazyspotify/lazyspotify-librespot`
- Arch `lazyspotify-bin`: `/usr/bin/lazyspotify` and `/usr/lib/lazyspotify/lazyspotify-librespot`

Example Linux package build:

```bash
go build -ldflags "-X github.com/dubeyKartikay/lazyspotify/buildinfo.PackagedDaemonPath=/usr/lib/lazyspotify/lazyspotify-librespot" -o target/lazyspotify ./cmd/lazyspotify
```

For Homebrew, inject `#{opt_libexec}/lazyspotify-librespot` during the formula
build so upgrades keep a stable absolute daemon path.

### GitHub Release archives

The Linux fallback assets install cleanly into package-managed locations.

The macOS fallback archive ships both binaries but does not provide relocatable
daemon discovery. After extracting it, set `librespot.daemon.cmd` explicitly in
your config.

### Source build

Build `lazyspotify`:

```bash
go build -o target/lazyspotify ./cmd/lazyspotify
```

Build the patched daemon from the forked `go-librespot` repository and set an
explicit config override:

```yaml
librespot:
  daemon:
    cmd:
      - /absolute/path/to/lazyspotify-librespot
```

## Version metadata

`lazyspotify version` and `lazyspotify --version` print:

- `version`
- `commit`
- `build_date`
- `packaged_daemon_path`

## Configuration

Config lives under the OS config directory:

- macOS: `~/Library/Application Support/lazyspotify/config.yaml`
- Linux: `~/.config/lazyspotify/config.yaml`

Only overrides are required. Package builds may provide a compiled default
daemon path.

### Set Your Spotify Client ID

`lazyspotify` now requires your own Spotify app client ID. Spotify does not
allow individual apps to use extended quota for other users, so the bundled
client ID has been removed.

Set `auth.client_id` in `config.yaml`:

```yaml
auth:
  client_id: your_spotify_app_client_id
```

Environment override:

```bash
export AUTH_CLIENT_ID=your_spotify_app_client_id
```

Example:

```yaml
auth:
  client_id: your_spotify_app_client_id
librespot:
  daemon:
    cmd:
      - /absolute/path/to/lazyspotify-librespot
```

## Development

Run the app:

```bash
make run
```

Build the app:

```bash
make build
```

Run tests:

```bash
go test ./...
```

Distribution packaging, release scripts, and CI workflow details live in
[`docs/distribution.md`](docs/distribution.md).
