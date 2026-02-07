# Release Setup Guide

To enable automated releases via GitHub Actions and Homebrew updates, you need to configure a Personal Access Token (PAT).

## 1. Generate a Personal Access Token (PAT)

1.  Go to **[GitHub Settings > Developer Settings > Personal Access Tokens (Classic)](https://github.com/settings/tokens/new)**.
2.  Select **"Generate new token (classic)"**.
    *   *Note: Use "Classic" to ensure compatibility with GoReleaser and Homebrew taps.*
3.  **Name**: `Needy Release Token`.
4.  **Expiration**: Set to your preference (e.g., "No expiration" or 90 days).
5.  **Select Scopes**:
    *   ✅ **`repo`** (Full control of private repositories).
        *   Required to create releases in `akafred/needy` and push formula updates to `akafred/homebrew-needy`.
    *   ✅ **`workflow`** (Optional, but good for Actions).
6.  Click **Generate token** and copy the value (starts with `ghp_`).

## 2. Add Secret to GitHub

1.  Navigate to the repository secrets:
    *   **Settings > Secrets and variables > Actions**
    *   Direct Link: `https://github.com/akafred/needy/settings/secrets/actions`
2.  Click **New repository secret**.
3.  **Name**: `GH_PAT`
    *   *Critical: This must match the name used in `.github/workflows/release.yml`.*
4.  **Secret**: Paste your token.
5.  Click **Add secret**.

## 3. Trigger a Release

Once the secret is configured, you can trigger a release by pushing a semantic version tag:

```bash
git tag v0.0.1
git push origin v0.0.1
```

The **Release** workflow will automatically:
1.  Build binaries for all platforms.
2.  Create a GitHub Release with changelog.
3.  Update the `akafred/homebrew-needy` tap.
