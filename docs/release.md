# Release

Release assets are built for:

- macOS arm64
- macOS amd64
- Linux 386
- Linux amd64
- Linux armv6
- Linux armv7
- Linux arm64
- Linux ppc64le
- Linux riscv64
- Linux s390x
- Windows 386
- Windows amd64
- Windows arm
- Windows arm64

The GitHub release workflow uses the same target names as the local `make build` command.

## Windows Signing

Tagged releases require Authenticode signing for every Windows `.exe` asset. The workflow signs Windows binaries with Azure Trusted Signing and verifies the signature before upload. If signing is not configured, the tagged release fails instead of publishing unsigned Windows binaries.

Required GitHub Actions secrets:

| Name | Purpose |
| --- | --- |
| `AZURE_CLIENT_ID` | Azure app registration client ID used by OIDC login. |
| `AZURE_TENANT_ID` | Azure tenant ID. |
| `AZURE_SUBSCRIPTION_ID` | Azure subscription that owns the signing account. |

Required GitHub Actions variables:

| Name | Purpose |
| --- | --- |
| `AZURE_TRUSTED_SIGNING_ENDPOINT` | Region endpoint, for example `https://eus.codesigning.azure.net`. |
| `AZURE_TRUSTED_SIGNING_ACCOUNT` | Azure Trusted Signing account name. |
| `AZURE_TRUSTED_SIGNING_CERT_PROFILE` | Certificate profile name. |

The Azure identity must have the `Artifact Signing Certificate Profile Signer` role on the certificate profile. The GitHub repository also needs a federated credential for OIDC login.

Windows release files are signed with SHA256 and timestamped through:

```text
http://timestamp.acs.microsoft.com
```

The release job also publishes `SHA256SUMS.txt` for all assets.
