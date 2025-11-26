<#
.SYNOPSIS
    Cross‑compile a Go project for the major OS/ARCH targets without
    hard‑coding a package path.

.DESCRIPTION
    The script builds the current package (or the package in the folder
    where you invoke it) for the following OS/ARCH combinations:

        OS: linux, darwin, windows, freebsd, openbsd, netbsd
        ARCH: amd64, arm64, 386, arm

    The binaries are stored in <OutputDir>/<os>_<arch>/.
    Windows binaries receive a .exe suffix automatically.

.PARAMETER OutputDir
    The directory where the compiled binaries will be placed.
    Default: "dist"

.EXAMPLE
    # Build everything into ./dist
    .\go-build-all.ps1

    # Build everything into ./build
    .\go-build-all.ps1 -OutputDir build
#>
param(
    [Parameter(Mandatory=$false)]
    [string]$OutputDir = "dist"
)

# ---- Configuration -------------------------------------------------
$OS_LIST   = @("linux","darwin","windows","freebsd","openbsd","netbsd")
$ARCH_LIST = @("amd64","arm64","386","arm")

# ---- Prepare output folder -----------------------------------------
if (-not (Test-Path $OutputDir)) { New-Item -ItemType Directory -Path $OutputDir | Out-Null }

Write-Host "Cross‑building for:"
Write-Host "  OSes : $($OS_LIST -join ', ')"
Write-Host "  Archs: $($ARCH_LIST -join ', ')"
Write-Host "  Output dir: $OutputDir`n"

# ---- Build loop -----------------------------------------------------
foreach ($os in $OS_LIST) {
    foreach ($arch in $ARCH_LIST) {
        # Create target folder and binary path
        $binPath = Join-Path $OutputDir ("$($script:AppName)")

        # Run the build in a subshell so env vars are local
        & {
            # Environment for this target
            $env:GOOS          = $os
            $env:GOARCH        = $arch
            $env:CGO_ENABLED   = 0
            $env:GOARM         = ($arch -eq 'arm') ? '7' : $null

            go build -o "$binPath/goUp-$os-$arch"

            Write-Host "  → Built: $binPath"
        }
    }
}

Write-Host "`nAll builds completed.  Binaries live in $OutputDir"
