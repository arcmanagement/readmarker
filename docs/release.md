# Release

This repo is wired for GitHub Releases and Homebrew Cask publishing through
GoReleaser. Publishing is intentionally tag-driven.

## One-time setup

Before the first public release:

1. Choose and add the repository license.
2. Create the public Homebrew tap repository:

   ```bash
   gh repo create arcmanagement/homebrew-readmarker \
     --public \
     --description "Homebrew tap for readmarker"
   ```

3. Create a fine-grained token that can write contents to
   `arcmanagement/homebrew-readmarker`, then store it on `arcmanagement/readmarker`:

   ```bash
   gh secret set HOMEBREW_TAP_GITHUB_TOKEN \
     --repo arcmanagement/readmarker \
     --body "$HOMEBREW_TAP_GITHUB_TOKEN"
   ```

4. Make the source repository public:

   ```bash
   gh repo edit arcmanagement/readmarker --visibility public
   ```

5. Apply the standard public-repo branch protection Rulesets after the visibility
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
keyless cosign, and publishes the Homebrew Cask to
`arcmanagement/homebrew-readmarker`.

## Verify

After the GitHub Actions release job passes:

```bash
gh release view v0.1.0 --repo arcmanagement/readmarker
brew update
brew install --cask arcmanagement/readmarker/readmarker
readmarker --version
```
