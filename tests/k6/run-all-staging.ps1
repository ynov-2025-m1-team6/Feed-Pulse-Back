# Script PowerShell pour exécuter tous les tests K6 sur Staging
Write-Host "🚀 Exécution de tous les tests K6 sur Staging" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Green
Write-Host ""

$tests = @(
    @{ Name = "Ping"; File = "ping.js" },
    @{ Name = "Login"; File = "login.js" },
    @{ Name = "Register"; File = "register.js" },
    @{ Name = "Auth Flow"; File = "auth.js" },
    @{ Name = "Feedback"; File = "feedback.js" },
    @{ Name = "Board"; File = "board.js" },
    @{ Name = "Full Flow"; File = "full-flow.js" }
)

$totalTests = $tests.Count
$currentTest = 1
$results = @()

foreach ($test in $tests) {
    Write-Host "📊 Test $currentTest/$totalTests : $($test.Name)" -ForegroundColor Cyan
    Write-Host "Fichier: $($test.File)" -ForegroundColor Gray
    
    $startTime = Get-Date
    
    try {
        $env:ENVIRONMENT = "staging"
        $output = k6 run $test.File 2>&1
        $exitCode = $LASTEXITCODE
        
        $endTime = Get-Date
        $duration = $endTime - $startTime
        
        if ($exitCode -eq 0) {
            Write-Host "✅ $($test.Name) - RÉUSSI" -ForegroundColor Green
            $status = "RÉUSSI"
        } else {
            Write-Host "⚠️ $($test.Name) - SEUILS DÉPASSÉS" -ForegroundColor Yellow
            $status = "SEUILS DÉPASSÉS"
        }
        
        $results += [PSCustomObject]@{
            Test = $test.Name
            Fichier = $test.File
            Status = $status
            Durée = $duration.ToString("mm\:ss")
            ExitCode = $exitCode
        }
        
    } catch {
        Write-Host "❌ $($test.Name) - ERREUR: $($_.Exception.Message)" -ForegroundColor Red
        $results += [PSCustomObject]@{
            Test = $test.Name
            Fichier = $test.File
            Status = "ERREUR"
            Durée = "N/A"
            ExitCode = -1
        }
    }
    
    Write-Host ""
    $currentTest++
}

Write-Host "🎉 Tous les tests sont terminés !" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Résumé des résultats:" -ForegroundColor Yellow
$results | Format-Table -AutoSize

Write-Host "📁 Consultez le dossier results/ pour les rapports détaillés" -ForegroundColor Cyan

# Compter les résultats
$reussis = ($results | Where-Object { $_.Status -eq "RÉUSSI" }).Count
$seuils = ($results | Where-Object { $_.Status -eq "SEUILS DÉPASSÉS" }).Count
$erreurs = ($results | Where-Object { $_.Status -eq "ERREUR" }).Count

Write-Host ""
Write-Host "📈 Statistiques finales:" -ForegroundColor Magenta
Write-Host "  ✅ Tests réussis: $reussis" -ForegroundColor Green
Write-Host "  ⚠️ Seuils dépassés: $seuils" -ForegroundColor Yellow  
Write-Host "  ❌ Erreurs: $erreurs" -ForegroundColor Red
Write-Host "  📊 Total: $totalTests tests" -ForegroundColor Cyan
