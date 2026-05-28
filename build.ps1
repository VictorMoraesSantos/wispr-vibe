# Wispr Vibe — Full Build Script
# This script builds everything from scratch:
#   1. Sets up whisper.cpp (prebuilt CUDA binary or source compile)
#   2. Downloads the whisper model
#   3. Builds the Go GUI application
#   4. Updates the config to point at the new binary
#
# GPU strategy (automatic, no CUDA Toolkit required):
#   - If CUDA Toolkit (nvcc) is found: compile whisper.cpp from source with CUDA
#   - Otherwise: download prebuilt CUDA-enabled binary from whisper.cpp releases
#     (includes all required CUDA runtime DLLs — self-contained)
#
# Prerequisites:
#   - Git
#   - CMake (only needed when compiling from source)
#   - Go 1.21+
#   - GCC (for Fyne/CGo) — e.g., via TDM-GCC or MSYS2
#
# Usage:
#   .\build.ps1              # Build everything (GPU enabled)
#   .\build.ps1 -NoGPU      # Build without GPU support
#   .\build.ps1 -SkipWhisper # Only rebuild the Go app
#   .\build.ps1 -Source      # Force compile from source (requires CUDA Toolkit)

param(
    [switch]$NoGPU,
    [switch]$SkipWhisper,
    [switch]$SkipModel,
    [switch]$Source,
    [string]$ModelSize = "small"
)

$ErrorActionPreference = "Stop"

$ROOT = $PSScriptRoot
$DEPS_DIR = Join-Path $ROOT "deps"
$WHISPER_DIR = Join-Path $DEPS_DIR "whisper.cpp"
$WHISPER_BUILD = Join-Path $WHISPER_DIR "build"
$OUTPUT_DIR = Join-Path $ROOT "dist"
$BIN_DIR = Join-Path $OUTPUT_DIR "bin"

$MODEL_DIR = Join-Path $OUTPUT_DIR "models"
$MODEL_MAP = @{
    "base"   = "ggml-base.bin"
    "small"  = "ggml-small.bin"
    "medium" = "ggml-medium.bin"
    "large"  = "ggml-large-v3-turbo.bin"
}
$MODEL_FILE = $MODEL_MAP[$ModelSize]
$MODEL_URL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/$MODEL_FILE"

function Write-Step($msg) {
    Write-Host "`n==> $msg" -ForegroundColor Cyan
}

function Test-Command($cmd) {
    return [bool](Get-Command $cmd -ErrorAction SilentlyContinue)
}

# Write JSON without BOM so Go's json.Unmarshal can parse it
function Write-JsonNoBOM($path, $obj) {
    $json = $obj | ConvertTo-Json -Depth 3
    [IO.File]::WriteAllText($path, $json, [Text.UTF8Encoding]::new($false))
}

function Update-ConfigValue($file, $key, $value) {
    if (-not (Test-Path $file)) { return }
    $cfg = Get-Content $file -Raw | ConvertFrom-Json
    $cfg.$key = $value
    Write-JsonNoBOM $file $cfg
}

# --- Validate prerequisites ---
Write-Step "Checking prerequisites"

if (-not (Test-Command "git")) { throw "Git not found. Install from https://git-scm.com" }
if (-not (Test-Command "go"))  { throw "Go not found. Install from https://go.dev/dl/" }

# Determine GPU strategy
$gpuEnabled = $false
$buildFromSource = $false

if (-not $NoGPU) {
    if ((Test-Command "nvcc") -or $Source) {
        if (Test-Command "nvcc") {
            $gpuEnabled = $true
            $buildFromSource = $true
            $nvccVersion = & nvcc --version 2>&1 | Select-String "release"
            Write-Host "  CUDA Toolkit found: $nvccVersion" -ForegroundColor Green
            Write-Host "  Strategy: compile from source with CUDA" -ForegroundColor Green
        } else {
            Write-Host "  -Source flag set but nvcc not found. Install CUDA Toolkit first." -ForegroundColor Red
            throw "nvcc not found"
        }
    } else {
        # No CUDA Toolkit: download prebuilt CUDA binary (includes all runtime DLLs)
        $gpuEnabled = $true
        $buildFromSource = $false
        Write-Host "  CUDA Toolkit not found — will download prebuilt CUDA binary" -ForegroundColor Cyan
        Write-Host "  (Includes all required CUDA runtime DLLs, no toolkit needed)" -ForegroundColor Cyan
    }
}

New-Item -ItemType Directory -Force -Path $BIN_DIR | Out-Null

# --- Setup whisper.cpp ---
if (-not $SkipWhisper) {
    if ($buildFromSource) {
        if (-not (Test-Command "cmake")) { throw "CMake not found. Install from https://cmake.org/download/" }

        Write-Step "Setting up whisper.cpp (source build)"

        New-Item -ItemType Directory -Force -Path $DEPS_DIR | Out-Null
        if (Test-Path $WHISPER_DIR) {
            Write-Host "  Updating existing whisper.cpp..."
            Push-Location $WHISPER_DIR; git pull --quiet; Pop-Location
        } else {
            Write-Host "  Cloning whisper.cpp..."
            git clone --depth 1 https://github.com/ggerganov/whisper.cpp.git $WHISPER_DIR
        }

        Write-Step "Compiling whisper.cpp with CUDA"
        if (Test-Path $WHISPER_BUILD) { Remove-Item -Recurse -Force $WHISPER_BUILD }
        New-Item -ItemType Directory -Force -Path $WHISPER_BUILD | Out-Null

        $cmakeArgs = @(
            "-B", $WHISPER_BUILD, "-S", $WHISPER_DIR,
            "-DCMAKE_BUILD_TYPE=Release",
            "-DWHISPER_BUILD_EXAMPLES=ON",
            "-DWHISPER_BUILD_TESTS=OFF",
            "-DGGML_CUDA=ON"
        )
        & cmake @cmakeArgs
        if ($LASTEXITCODE -ne 0) { throw "CMake configuration failed" }
        & cmake --build $WHISPER_BUILD --config Release --parallel
        if ($LASTEXITCODE -ne 0) { throw "whisper.cpp build failed" }

        $whisperExe = Get-ChildItem -Path $WHISPER_BUILD -Recurse -Filter "whisper-cli.exe" | Select-Object -First 1
        if (-not $whisperExe) { throw "whisper-cli.exe not found in build output" }
        Copy-Item $whisperExe.FullName (Join-Path $BIN_DIR "whisper-cli.exe") -Force
        # Copy all DLLs (ggml-cuda.dll, cudart, etc.)
        $dllDir = Split-Path $whisperExe.FullName
        Get-ChildItem -Path $dllDir -Filter "*.dll" | ForEach-Object { Copy-Item $_.FullName $BIN_DIR -Force }
        Get-ChildItem -Path $WHISPER_BUILD -Recurse -Filter "ggml*.dll" -ErrorAction SilentlyContinue | ForEach-Object {
            Copy-Item $_.FullName $BIN_DIR -Force
        }
        Write-Host "  Built with CUDA: $BIN_DIR\whisper-cli.exe" -ForegroundColor Green

    } elseif (-not $NoGPU) {
        # Download prebuilt CUDA binary from whisper.cpp releases
        Write-Step "Downloading prebuilt CUDA whisper binary"
        Write-Host "  Fetching latest release info..."
        $ProgressPreference = 'SilentlyContinue'

        $release = Invoke-RestMethod "https://api.github.com/repos/ggerganov/whisper.cpp/releases/latest" -UseBasicParsing
        $cudaAsset = $release.assets | Where-Object { $_.name -match "cublas-12" }
        if (-not $cudaAsset) {
            $cudaAsset = $release.assets | Where-Object { $_.name -match "cublas-11" }
        }
        if (-not $cudaAsset) {
            Write-Host "  No CUDA prebuilt binary found in release. Falling back to CPU." -ForegroundColor Yellow
            $gpuEnabled = $false
        } else {
            $zipPath = Join-Path $OUTPUT_DIR "whisper-cublas.zip"
            $sizeMB = [math]::Round($cudaAsset.size / 1MB, 0)
            Write-Host "  Downloading $($cudaAsset.name) ($sizeMB MB)..." -ForegroundColor Cyan
            Write-Host "  (First run only — this may take several minutes)" -ForegroundColor Gray

            if (Test-Path $zipPath) { Remove-Item $zipPath -Force }
            Invoke-WebRequest $cudaAsset.browser_download_url -OutFile $zipPath -UseBasicParsing
            $ProgressPreference = 'Continue'

            Write-Host "  Extracting..."
            Add-Type -AssemblyName System.IO.Compression.FileSystem
            $zip = [IO.Compression.ZipFile]::OpenRead($zipPath)
            foreach ($entry in $zip.Entries) {
                $destPath = Join-Path $BIN_DIR $entry.Name
                $stream = $entry.Open()
                $fs = [IO.File]::Create($destPath)
                $stream.CopyTo($fs); $fs.Close(); $stream.Close()
            }
            $zip.Dispose()

            $hasCudaDLL = Test-Path (Join-Path $BIN_DIR "ggml-cuda.dll")
            if ($hasCudaDLL) {
                Write-Host "  CUDA binary ready (ggml-cuda.dll present)" -ForegroundColor Green
            } else {
                Write-Host "  Warning: ggml-cuda.dll not found in extracted files" -ForegroundColor Yellow
                $gpuEnabled = $false
            }
        }

    } else {
        # CPU-only: download plain prebuilt binary
        Write-Step "Downloading prebuilt CPU whisper binary"
        $ProgressPreference = 'SilentlyContinue'
        $release = Invoke-RestMethod "https://api.github.com/repos/ggerganov/whisper.cpp/releases/latest" -UseBasicParsing
        $asset = $release.assets | Where-Object { $_.name -eq "whisper-bin-x64.zip" }
        $zipPath = Join-Path $OUTPUT_DIR "whisper-cpu.zip"
        Write-Host "  Downloading $($asset.name)..."
        Invoke-WebRequest $asset.browser_download_url -OutFile $zipPath -UseBasicParsing
        $ProgressPreference = 'Continue'
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        $zip = [IO.Compression.ZipFile]::OpenRead($zipPath)
        foreach ($entry in $zip.Entries) {
            $destPath = Join-Path $BIN_DIR $entry.Name
            $stream = $entry.Open(); $fs = [IO.File]::Create($destPath)
            $stream.CopyTo($fs); $fs.Close(); $stream.Close()
        }
        $zip.Dispose()
        Write-Host "  CPU binary ready" -ForegroundColor Green
    }
}

# --- Copy/link model ---
# Check if the user already has models in ~/.wispr-vibe/models
$userModelDir = Join-Path $env:USERPROFILE ".wispr-vibe\models"
$distModelPath = Join-Path $BIN_DIR $MODEL_FILE
$userModelPath = Join-Path $userModelDir $MODEL_FILE

# --- Download model ---
if (-not $SkipModel) {
    Write-Step "Setting up whisper model ($ModelSize)"
    New-Item -ItemType Directory -Force -Path $BIN_DIR | Out-Null

    if (Test-Path $distModelPath) {
        Write-Host "  Model already at dist: $MODEL_FILE" -ForegroundColor Green
    } elseif (Test-Path $userModelPath) {
        Write-Host "  Linking existing model from ~/.wispr-vibe/models/$MODEL_FILE" -ForegroundColor Green
        Copy-Item $userModelPath $distModelPath
    } else {
        Write-Host "  Downloading $MODEL_FILE (this may take a few minutes)..."
        try {
            $ProgressPreference = 'SilentlyContinue'
            Invoke-WebRequest -Uri $MODEL_URL -OutFile $distModelPath -UseBasicParsing
            $ProgressPreference = 'Continue'
            Write-Host "  Downloaded: $MODEL_FILE" -ForegroundColor Green
        } catch {
            Write-Host "  Download failed. Download manually:" -ForegroundColor Red
            Write-Host "  $MODEL_URL -> $distModelPath" -ForegroundColor Yellow
        }
    }
}

# Resolve final model path (prefer dist, fallback to user dir)
$finalModelPath = if (Test-Path $distModelPath) { $distModelPath } else { $userModelPath }

# --- Build Go application ---
Write-Step "Building Wispr Vibe GUI"

$env:CGO_ENABLED = "1"
Push-Location $ROOT
go build -ldflags="-s -w" -o (Join-Path $BIN_DIR "vibevoice-gui.exe") ./cmd/vibevoice-gui/
if ($LASTEXITCODE -ne 0) { throw "Go build failed" }
Pop-Location
Write-Host "  Built: vibevoice-gui.exe" -ForegroundColor Green

# --- Always update config with the new binary path ---
Write-Step "Updating config"

$configDir = Join-Path $env:USERPROFILE ".wispr-vibe"
$configFile = Join-Path $configDir "config.json"
New-Item -ItemType Directory -Force -Path $configDir | Out-Null

if (-not (Test-Path $configFile)) {
    $defaultConfig = [ordered]@{
        stt_engine         = "whisper_local"
        provider           = "local"
        whisper_api_key    = "not-needed"
        whisper_exe_path   = (Join-Path $BIN_DIR "whisper-cli.exe")
        whisper_model_path = $finalModelPath
        sample_rate        = 16000
        language           = ""
        log_level          = "info"
        hotkey             = "Ctrl+Shift+R"
        remove_fillers     = $true
        fix_punctuation    = $true
        use_gpu            = $gpuEnabled
    }
    $defaultConfig | ConvertTo-Json -Depth 2 | ForEach-Object { [IO.File]::WriteAllText($configFile, $_, [Text.UTF8Encoding]::new($false)) }
    Write-Host "  Config created: $configFile" -ForegroundColor Green
} else {
    # Update the binary path and GPU flag in the existing config
    $cfg = Get-Content $configFile -Raw | ConvertFrom-Json
    $cfg.whisper_exe_path = (Join-Path $BIN_DIR "whisper-cli.exe")
    $cfg.use_gpu = $gpuEnabled
    if ($finalModelPath -and (Test-Path $finalModelPath)) {
        $cfg.whisper_model_path = $finalModelPath
    }
    Write-JsonNoBOM $configFile $cfg
    Write-Host "  Config updated: whisper_exe_path, use_gpu, whisper_model_path" -ForegroundColor Green
}

# --- Done ---
Write-Step "Build complete!"
Write-Host ""
Write-Host "  Output: $BIN_DIR" -ForegroundColor White
Write-Host "  Run:    $BIN_DIR\vibevoice-gui.exe" -ForegroundColor White
Write-Host ""

if ($gpuEnabled) {
    $cudaFlag = if (Test-Path (Join-Path $BIN_DIR "ggml-cuda.dll")) { "ggml-cuda.dll present" } else { "source build" }
    Write-Host "  GPU: CUDA enabled ($cudaFlag)" -ForegroundColor Green
} else {
    Write-Host "  GPU: Disabled (CPU-only transcription)" -ForegroundColor Yellow
}
Write-Host ""
