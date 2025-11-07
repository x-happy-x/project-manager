param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Args
)

$enginePath = Join-Path -Path $PSScriptRoot -ChildPath '..\dist\pm-engine.exe'
$enginePath = Resolve-Path -LiteralPath $enginePath | Select-Object -ExpandProperty Path

if ($Args.Count -eq 0 -or $Args[0] -in @('ls','add','rm','-h','--help')) {
    & $enginePath @Args
    return
}

$dialect  = $env:PM_DIALECT
if (-not $dialect) { $dialect = 'pwsh' }

$script   = & $enginePath --dialect $dialect @Args

if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

if ($null -ne $script -and $script -ne "") {
    $cmd = ($script -join "`n")
    Invoke-Expression -Command $cmd
}
