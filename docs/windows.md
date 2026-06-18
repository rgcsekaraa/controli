# Windows

Windows is supported as a guest platform.

The default join mode uses the browser terminal. This avoids native console rendering limits during large terminal redraws.

## Device Guard and Smart App Control

Windows Device Guard, App Control for Business, and Smart App Control can block unsigned executables. Controli tagged releases are configured to require Authenticode signing for Windows `.exe` assets before they are uploaded.

If a company-managed Windows machine still blocks the signed executable, the organization's App Control policy must allow the Controli publisher or release hash. A local unblock command cannot override a strict organization policy.

Recommended command:

```powershell
.\controli.exe join 1234567
```
