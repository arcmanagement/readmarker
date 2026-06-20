# Release

This repo is wired for GitHub Releases and Homebrew formula publishing through
GoReleaser. Publishing is intentionally tag-driven.

## One-time setup

Before the first public release:

1. Choose and add the repository license.
2. Make the source repository public:

   ```bash
   gh repo edit arcmanagement/readmarker --visibility public
   ```

3. Apply the standard public-repo branch protection Rulesets after the visibility
   change. Do a dry-run first, then have the owner run the write step.

## Release

Create and push a signed version tag:

```bash
git switch main
git pull --ff-only origin main
git tag -s v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

The tag starts `.github/workflows/release.yml`, which runs GoReleaser. The
release job builds `darwin` and `linux` binaries for `amd64` and `arm64`,
uploads release archives, writes `checksums.txt`, signs the checksum file with
keyless cosign, and opens a pull request that updates this repository's
`Formula/readmarker.rb`.

Merge the generated formula pull request after checking that it points to the
new release assets.

## Verify

After the GitHub Actions release job passes:

```bash
gh release view v0.1.0 --repo arcmanagement/readmarker
brew update
brew untap arcmanagement/readmarker
brew tap arcmanagement/readmarker https://github.com/arcmanagement/readmarker
brew trust --formula arcmanagement/readmarker/readmarker
brew install readmarker
readmarker --version
```
