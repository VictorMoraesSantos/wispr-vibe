# Wispr Vibe — Full Build Script (with CUDA-enabled whisper.cpp)
# This script builds everything from scratch:
#   1. Clones and compiles whisper.cpp with CUDA support
#   2. Downloads the whisper model
#   3. Builds the Go GUI application
#
# Prerequisites:
#   - Git
#   - CMake (3.14+)
#   - Visual Studio Build Tools (C++ workload) or MinGW with CUDA
#   - NVIDIA CUDA Toolkit (https://developer.nvidia.com/cuda-downloads)
#   - Go 1.21+
#   - GCC (for Fyne/CGo)
#
# Usage:
#   .\build.ps1              # Build everything (GPU enabled)
#   .\build.ps1 -NoGPU      # Build without GPU support
#   .\build.ps1 -SkipWhisper # Only rebuild the Go app

param(
    [switch]$NoGPU,
    [switch]$SkipWhisper,
    [switch]$SkipModel,
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

# --- Validate prerequisites ---
Write-Step "Checking prerequisites"

if (-not (Test-Command "git")) { throw "Git not found. Install from https://git-scm.com" }
if (-not (Test-Command "go"))  { throw "Go not found. Install from https://go.dev/dl/" }
if (-not (Test-Command "cmake")) { throw "CMake not found. Install from https://cmake.org/download/" }

$cudaAvailable = $false
if (-not $NoGPU) {
    if (Test-Command "nvcc") {
        $cudaAvailable = $true
        $nvccVersion = & nvcc --version 2>&1 | Select-String "release"
        Write-Host "  CUDA found: $nvccVersion" -ForegroundColor Green
    } else {
        Write-Host "  CUDA not found. Building CPU-only." -ForegroundColor Yellow
        Write-Host "  Install CUDA Toolkit: https://developer.nvidia.com/cuda-downloads" -ForegroundColor Yellow
        $NoGPU = $true
    }
}

# --- Clone/update whisper.cpp ---
if (-not $SkipWhisper) {
    Write-Step "Setting up whisper.cpp"

    New-Item -ItemType Directory -Force -Path $DEPS_DIR | Out-Null

    if (Test-Path $WHISPER_DIR) {
        Write-Host "  Updating existing whisper.cpp..."
        Push-Location $WHISPER_DIR
        git pull --quiet
        Pop-Location
    } else {
        Write-Host "  Cloning whisper.cpp..."
        git clone --depth 1 https://github.com/ggerganov/whisper.cpp.git $WHISPER_DIR
    }

    # --- Build whisper.cpp ---
    Write-Step "Building whisper.cpp $(if ($NoGPU) {'(CPU)'} else {'(CUDA)'})"

    if (Test-Path $WHISPER_BUILD) {
        Remove-Item -Recurse -Force $WHISPER_BUILD
    }
    New-Item -ItemType Directory -Force -Path $WHISPER_BUILD | Out-Null

    $cmakeArgs = @(
        "-B", $WHISPER_BUILD,
        "-S", $WHISPER_DIR,
        "-DCMAKE_BUILD_TYPE=Release",
        "-DWHISPER_BUILD_EXAMPLES=ON",
        "-DWHISPER_BUILD_TESTS=OFF"
    )

    if (-not $NoGPU -and $cudaAvailable) {
        $cmakeArgs += "-DGGML_CUDA=ON"
        Write-Host "  GPU acceleration: ENABLED" -ForegroundColor Green
    } else {
        Write-Host "  GPU acceleration: DISABLED (CPU only)" -ForegroundColor Yellow
    }

    & cmake @cmakeArgs
    if ($LASTEXITCODE -ne 0) { throw "CMake configuration failed" }

    & cmake --build $WHISPER_BUILD --config Release --parallel
    if ($LASTEXITCODE -ne 0) { throw "whisper.cpp build failed" }

    # --- Copy whisper binary ---
    New-Item -ItemType Directory -Force -Path $BIN_DIR | Out-Null

    $whisperExe = Get-ChildItem -Path $WHISPER_BUILD -Recurse -Filter "whisper-cli.exe" | Select-Object -First 1
    if (-not $whisperExe) {
        $whisperExe = Get-ChildItem -Path $WHISPER_BUILD -Recurse -Filter "main.exe" | Select-Object -First 1
    }
    if ($whisperExe) {
        Copy-Item $whisperExe.FullName (Join-Path $BIN_DIR "whisper-cli.exe") -Force
        Write-Host "  Copied: whisper-cli.exe" -ForegroundColor Green

        # Copy CUDA DLLs if GPU build
        if (-not $NoGPU -and $cudaAvailable) {
            $dllDir = Split-Path $whisperExe.FullName
            Get-ChildItem -Path $dllDir -Filter "*.dll" -ErrorAction SilentlyContinue | ForEach-Object {
                Copy-Item $_.FullName $BIN_DIR -Force
            }
            # Also copy ggml shared libs
            Get-ChildItem -Path $WHISPER_BUILD -Recurse -Filter "ggml*.dll" -ErrorAction SilentlyContinue | ForEach-Object {
                Copy-Item $_.FullName $BIN_DIR -Force
            }
        }
    } else {
        throw "whisper-cli.exe not found in build output"
    }
}

# --- Download model ---
if (-not $SkipModel) {
    Write-Step "Downloading whisper model ($ModelSize)"

    New-Item -ItemType Directory -Force -Path $MODEL_DIR | Out-Null
    $modelPath = Join-Path $MODEL_DIR $MODEL_FILE

    if (Test-Path $modelPath) {
        Write-Host "  Model already exists: $MODEL_FILE" -ForegroundColor Green
    } else {
        Write-Host "  Downloading $MODEL_FILE (this may take a few minutes)..."
        try {
            $ProgressPreference = 'SilentlyContinue'
            Invoke-WebRequest -Uri $MODEL_URL -OutFile $modelPath -UseBasicParsing
            $ProgressPreference = 'Continue'
            Write-Host "  Downloaded: $MODEL_FILE" -ForegroundColor Green
        } catch {
            Write-Host "  Download failed. Download manually from:" -ForegroundColor Red
            Write-Host "  $MODEL_URL" -ForegroundColor Yellow
            Write-Host "  Place in: $MODEL_DIR" -ForegroundColor Yellow
        }
    }
}

# --- Build Go application ---
Write-Step "Building Wispr Vibe GUI"

$env:CGO_ENABLED = "1"
Push-Location $ROOT
go build -ldflags="-s -w" -o (Join-Path $BIN_DIR "vibevoice-gui.exe") ./cmd/vibevoice-gui/
if ($LASTEXITCODE -ne 0) { throw "Go build failed" }
Pop-Location
Write-Host "  Built: vibevoice-gui.exe" -ForegroundColor Green

# --- Create default config ---
Write-Step "Creating default config"

$configDir = Join-Path $env:USERPROFILE ".wispr-vibe"
$configFile = Join-Path $configDir "config.json"

if (-not (Test-Path $configFile)) {
    New-Item -ItemType Directory -Force -Path $configDir | Out-Null

    $defaultConfig = @{
        stt_engine        = "whisper_local"
        provider          = "local"
        whisper_api_key   = "not-needed"
        whisper_exe_path  = (Join-Path $BIN_DIR "whisper-cli.exe")
        whisper_model_path = (Join-Path $MODEL_DIR $MODEL_FILE)
        sample_rate       = 16000
        language          = ""
        log_level         = "info"
        hotkey            = "Ctrl+Shift+R"
        remove_fillers    = $true
        fix_punctuation   = $true
        use_gpu           = (-not $NoGPU -and $cudaAvailable)
    } | ConvertTo-Json -Depth 2

    Set-Content -Path $configFile -Value $defaultConfig -Encoding UTF8
    Write-Host "  Config created: $configFile" -ForegroundColor Green
} else {
    Write-Host "  Config already exists, skipping" -ForegroundColor Yellow
}

# --- Done ---
Write-Step "Build complete!"
Write-Host ""
Write-Host "  Output: $BIN_DIR" -ForegroundColor White
Write-Host "  Run:    $BIN_DIR\vibevoice-gui.exe" -ForegroundColor White
Write-Host ""

if (-not $NoGPU -and $cudaAvailable) {
    Write-Host "  GPU: CUDA enabled (transcription will use your NVIDIA GPU)" -ForegroundColor Green
} else {
    Write-Host "  GPU: Disabled (CPU-only transcription)" -ForegroundColor Yellow
}
Write-Host ""
