# Windows through WSL2

Jin supports macOS and Linux, on amd64 and arm64. On Windows, use the Linux build
inside WSL2. Native Windows binaries, Git Bash integration, a PowerShell installer,
and Windows-specific self-update are not supported.

## Install

Install WSL2 and a Linux distribution, such as Ubuntu. If WSL is not installed,
run this in an administrator PowerShell terminal and follow Windows' restart prompts:

```powershell
wsl --install
```

Open the distribution in Windows Terminal. Run the normal Linux installer there:

```sh
curl -fsSL https://raw.githubusercontent.com/ethanhamilthon/jin/main/install.sh | sh
```

The installer selects the architecture of the Linux environment. Install your
project's tools, Git, and dependencies inside that environment too. `jin update`
also runs inside WSL and uses the Linux release archive.

## Projects

Prefer projects under the Linux home directory, such as `~/projects/my-app`, rather
than `/mnt/c/...`. This avoids Windows-mounted filesystem performance and file
permission differences. Start jin from that project's directory:

```sh
cd ~/projects/my-app
jin
```

Commands, hooks, background tasks, and editors run inside Linux. Jin stores its data
in the Linux home directory's `~/.jin`, not the Windows user profile. Windows Terminal
provides the terminal for the TUI; headless commands work as on other Linux systems.

## Web UI

Run inside WSL:

```sh
jin web --no-open
```

Copy the full localhost URL jin prints, including its token, into the Windows browser.
WSL2 normally forwards localhost connections to the Linux environment. If the URL does
not connect, check WSL networking, localhost forwarding, and firewall settings. Do not
expose jin publicly or disable its authentication to work around a networking problem.

## Validation

The WSL route uses the existing Linux implementation, not a separate Windows backend.
This workflow has not yet been checked in WSL2 for the current changes. The earlier
native Windows VM check does not validate WSL2. See Microsoft's setup documentation:
https://learn.microsoft.com/windows/wsl/install
