param (
    [string]$JWTSecret = "testjwtsecret",
    [string]$ServerPort = "8080"
)

$ErrorActionPreference = "Stop"
$BASE_URL = "http://localhost:$ServerPort"
$TIMESTAMP = Get-Date -Format "yyyy-MM-dd_HH-mm-ss"
$TEST_OUTPUT_DIR = "results/$TIMESTAMP"

Write-Host "Starting LeetCodeDuels Stress Test Pipeline" -ForegroundColor Green

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "..\..\..") # Go up to project root
$ServerDir = Join-Path $ProjectRoot "server/cmd/server"
$K6TestFile = Join-Path $ScriptDir "k6.js"
$DockerComposeFile = Join-Path $ScriptDir "docker-compose.test.yml"

Write-Host "Script Directory: $ScriptDir" -ForegroundColor Cyan
Write-Host "Project Root: $ProjectRoot" -ForegroundColor Cyan
Write-Host "Server Directory: $ServerDir" -ForegroundColor Cyan
Write-Host "K6 Test File: $K6TestFile" -ForegroundColor Cyan

# Create test output directory
if (!(Test-Path $TEST_OUTPUT_DIR)) {
    New-Item -ItemType Directory -Path $TEST_OUTPUT_DIR
}

function Cleanup {
    Write-Host "Cleaning up..." -ForegroundColor Yellow
    if ($serverProcess -and !$serverProcess.HasExited) {
        Write-Host "Stopping server process..." -ForegroundColor Yellow
        # Find child 'go' process and kill it, then the cmd process
        $childProcesses = Get-CimInstance Win32_Process | Where-Object { $_.ParentProcessId -eq $serverProcess.Id }
        foreach ($child in $childProcesses) {
            Stop-Process -Id $child.ProcessId -Force
        }
        $serverProcess.Kill()
        $serverProcess.WaitForExit(5000)
    }
    Write-Host "Stopping Docker containers..." -ForegroundColor Yellow
    docker-compose -f $DockerComposeFile down -v
}

trap { Cleanup; exit 1 }

try {
    if (!(Test-Path $K6TestFile)) {
        throw "K6 test file not found: $K6TestFile"
    }
    if (!(Test-Path $DockerComposeFile)) {
        throw "Docker Compose file not found: $DockerComposeFile"
    }
    if (!(Test-Path $ServerDir)) {
        throw "Server directory not found: $ServerDir"
    }

    Write-Host "Starting test services with Docker Compose..." -ForegroundColor Blue
    Push-Location $ScriptDir
    docker-compose -f docker-compose.test.yml up -d
    Pop-Location

    Write-Host "Waiting for services to be healthy..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
    
    # Check container health
    $maxHealthRetries = 10
    $healthRetryCount = 0
    do {
        $healthRetryCount++
        $containers = docker-compose -f $DockerComposeFile ps --format json | ConvertFrom-Json
        $allHealthy = $true
        
        foreach ($container in $containers) {
            if ($container.Health -and $container.Health -ne "healthy") {
                $allHealthy = $false
                break
            }
        }
        
        if ($allHealthy) {
            Write-Host "All services are healthy!" -ForegroundColor Green
            break
        }
        
        if ($healthRetryCount -ge $maxHealthRetries) {
            throw "Services failed to become healthy after $maxHealthRetries attempts"
        }
        
        Write-Host "Waiting for services to be healthy... ($healthRetryCount/$maxHealthRetries)" -ForegroundColor Yellow
        Start-Sleep -Seconds 2
    } while ($true)
    
    $env:DB_URL = "postgresql://testuser@localhost:5433/leetcodeduels_test?sslmode=disable"
    $env:RDB_URL = "redis://localhost:6379"
    $env:JWT_SECRET = $JWTSecret
    $env:PORT = $ServerPort
    $env:LOG_LEVEL = "error"
    $env:SUBMISSION_VALIDATION = "disable"

    Write-Host "Creating database schema..." -ForegroundColor Blue
    $migrationsDir = Join-Path $ProjectRoot "server\tests\migrations"
    
    if (Test-Path $migrationsDir) {
        $schemaFile = Join-Path $migrationsDir "0001_create_schema.up.sql"
        if (Test-Path $schemaFile) {
            Write-Host "Running schema migration..." -ForegroundColor Yellow
            $schemaContent = Get-Content $schemaFile -Raw
            docker exec k6-postgres-test-1 psql -U testuser -d leetcodeduels_test -c $schemaContent
            if ($LASTEXITCODE -ne 0) {
                throw "Schema creation failed"
            }
            Write-Host "Schema created successfully!" -ForegroundColor Green
        }
        
        $seedFile = Join-Path $migrationsDir "0002_seed_data.up.sql"
        if (Test-Path $seedFile) {
            Write-Host "Running seed data migration..." -ForegroundColor Yellow
            $seedContent = Get-Content $seedFile -Raw
            docker exec k6-postgres-test-1 psql -U testuser -d leetcodeduels_test -c $seedContent
            if ($LASTEXITCODE -ne 0) {
                throw "Seed data insertion failed"
            }
            Write-Host "Seed data inserted successfully!" -ForegroundColor Green
        }
    } else {
        Write-Host "Warning: Migrations directory not found at $migrationsDir" -ForegroundColor Yellow
    }

    Write-Host "Starting server from: $ServerDir" -ForegroundColor Blue
    $serverOutFile = Join-Path $TEST_OUTPUT_DIR "server-output-$TIMESTAMP.log"
    $serverErrFile = Join-Path $TEST_OUTPUT_DIR "server-error-$TIMESTAMP.log"

    $serverProcess = Start-Process -FilePath "go" -ArgumentList "run", "main.go" `
        -WorkingDirectory $ServerDir `
        -RedirectStandardOutput $serverOutFile `
        -RedirectStandardError $serverErrFile `
        -NoNewWindow -PassThru

    Write-Host "Waiting for server to be ready..." -ForegroundColor Yellow
    $maxRetries = 10
    $retryCount = 0
    do {
        $retryCount++
        try {
            $response = Invoke-WebRequest -Uri "$BASE_URL/api/v1/health" -TimeoutSec 2 -ErrorAction Stop
            if ($response.StatusCode -eq 200) {
                Write-Host "Server is ready!" -ForegroundColor Green
                break
            }
        } catch {
            # Silently continue retrying
        }
        
        if ($retryCount -ge $maxRetries) {
            throw "Server failed to start after $maxRetries attempts"
        }
        Start-Sleep -Seconds 2
    } while ($true)

    Write-Host "Running K6 stress tests..." -ForegroundColor Blue
    $k6MetricsFile = Join-Path $TEST_OUTPUT_DIR "k6-metrics.json"
    $k6SummaryFile = Join-Path $TEST_OUTPUT_DIR "k6-summary.txt"
    $k6ConsoleLogFile = Join-Path $TEST_OUTPUT_DIR "k6-console.log"
    
    $env:BASE_URL = $BASE_URL
    
    Push-Location $ScriptDir
    $k6Process = Start-Process -FilePath "k6" -ArgumentList @(
        "run",
        "--out", "json=$k6MetricsFile",
        "k6.js"
    ) -RedirectStandardOutput $k6SummaryFile -RedirectStandardError $k6ConsoleLogFile -NoNewWindow -PassThru -Wait
    Pop-Location

    if ($k6Process.ExitCode -eq 0) {
        Write-Host "K6 tests completed successfully!" -ForegroundColor Green
    } else {
        Write-Host "K6 tests failed with exit code: $($k6Process.ExitCode)" -ForegroundColor Red
        
        if (Test-Path $k6ConsoleLogFile) {
            Write-Host "Console log output from k6:" -ForegroundColor Red
            Get-Content $k6ConsoleLogFile | Write-Host -ForegroundColor Red
        }
    }

    Write-Host "Generating test summary..." -ForegroundColor Blue
    $summaryFile = Join-Path $TEST_OUTPUT_DIR "test-summary.txt"
    
    @"
LeetCodeDuels Stress Test Results
=================================
Timestamp: $TIMESTAMP
Base URL: $BASE_URL
K6 Exit Code: $($k6Process.ExitCode)

Files Generated in: $TEST_OUTPUT_DIR
- Server Output Log: $(Split-Path $serverOutFile -Leaf)
- Server Error Log:  $(Split-Path $serverErrFile -Leaf)
- K6 Final Summary:  $(Split-Path $k6SummaryFile -Leaf)
- K6 Console Log:    $(Split-Path $k6ConsoleLogFile -Leaf)
- K6 Raw Metrics:    $(Split-Path $k6MetricsFile -Leaf)

Test Configuration:
- Server Port: $ServerPort
"@ | Out-File -FilePath $summaryFile -Encoding UTF8

    Write-Host "Results saved to: $TEST_OUTPUT_DIR" -ForegroundColor Cyan
    Write-Host "Summary Report: $summaryFile" -ForegroundColor Cyan
    Write-Host "Server Output Log: $serverOutFile" -ForegroundColor Cyan
    Write-Host "Server Error Log: $serverErrFile" -ForegroundColor Cyan
    Write-Host "K6 Final Summary: $k6SummaryFile" -ForegroundColor Cyan
    Write-Host "K6 Console Log: $k6ConsoleLogFile" -ForegroundColor Cyan
    Write-Host "K6 Raw Metrics: $k6MetricsFile" -ForegroundColor Cyan

} catch {
    Write-Host "Error occurred: $($_.Exception.Message)" -ForegroundColor Red
    throw
} finally {
    Cleanup
    Write-Host "Pipeline completed!" -ForegroundColor Green
}
