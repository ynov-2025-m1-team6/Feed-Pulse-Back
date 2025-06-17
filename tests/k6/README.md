# Tests K6 pour Feed-Pulse-Back

Ce dossier contient une suite complète de tests de performance K6 pour l'API Feed-Pulse-Back.

## 📁 Structure des Tests

### Tests Individuels

- **`ping.js`** - Test de santé de l'API (endpoint `/ping`)
- **`login.js`** - Test de performance pour l'authentification
- **`register.js`** - Test d'inscription des utilisateurs
- **`auth.js`** - Test complet du flux d'authentification (login → user info → logout)
- **`feedback.js`** - Test des endpoints de feedback (upload, fetch, analyses)
- **`board.js`** - Test des métriques du tableau de bord
- **`full-flow.js`** - Test du flux complet de l'application
- **`stress-test.js`** - Test de stress avec charge élevée

### Scripts d'Exécution

- **`run-tests.ps1`** - Script PowerShell pour exécuter les tests facilement
- **`package.json`** - Configuration npm avec scripts de test

## 🚀 Utilisation

### Prérequis

1. **K6 installé** : [Installation K6](https://k6.io/docs/getting-started/installation/)
2. **Node.js** (pour les scripts npm)
3. **Go** (pour démarrer le serveur local)

### Méthode 1 : Script PowerShell (Recommandé)

```powershell
# Tests basiques en local
.\tests\k6\run-tests.ps1

# Tous les tests en local
.\tests\k6\run-tests.ps1 -TestType all

# Test spécifique
.\tests\k6\run-tests.ps1 -TestType feedback

# Tests sur staging
.\tests\k6\run-tests.ps1 -Environment staging

# Test de stress sur production
.\tests\k6\run-tests.ps1 -TestType stress -Environment prod

# Aide
.\tests\k6\run-tests.ps1 -Help
```

### Méthode 2 : Scripts NPM

```bash
cd tests/k6

# Tests basiques
npm test

# Tests individuels
npm run test:ping
npm run test:login
npm run test:register
npm run test:auth
npm run test:feedback
npm run test:board
npm run test:full-flow
npm run test:stress

# Tous les tests
npm run test:all

# Tests sur différents environnements
npm run test:staging
npm run test:prod
npm run test:staging:stress
npm run test:prod:stress
```

### Méthode 3 : K6 Direct

```bash
cd tests/k6

# Test local
k6 run ping.js
k6 run login.js

# Test avec environnement
k6 run -e ENVIRONMENT=staging auth.js
k6 run -e ENVIRONMENT=prod feedback.js
```

## 🌍 Environnements

### Local
- **URL** : `http://localhost:3000`
- **Usage** : Développement et tests locaux

### Staging
- **URL** : `https://feed-pulse-api-dev.onrender.com`
- **Usage** : Tests avant déploiement en production

### Production
- **URL** : `https://feed-pulse-api.onrender.com`
- **Usage** : Tests de performance en production

## 📊 Types de Tests

### 1. Tests de Santé (`ping.js`)
- Vérifie que l'API répond
- Temps de réponse < 200ms
- Charge : 50-200 utilisateurs virtuels

### 2. Tests d'Authentification (`login.js`, `auth.js`)
- Test de login avec utilisateur valide
- Vérification des tokens d'autorisation
- Flux complet : login → user info → logout
- Charge : 15-50 utilisateurs virtuels

### 3. Tests d'Inscription (`register.js`)
- Création de nouveaux utilisateurs
- Gestion des doublons
- Charge : 10-20 utilisateurs virtuels

### 4. Tests de Feedback (`feedback.js`)
- Upload de fichiers CSV
- Récupération des feedbacks
- Analyses par utilisateur
- Charge : 10-15 utilisateurs virtuels

### 5. Tests de Tableau de Bord (`board.js`)
- Métriques du dashboard
- Performance des requêtes d'agrégation
- Charge : 15-30 utilisateurs virtuels

### 6. Tests de Flux Complet (`full-flow.js`)
- Simulation d'un parcours utilisateur complet
- De la connexion à la déconnexion
- Toutes les fonctionnalités principales
- Charge : 8-12 utilisateurs virtuels

### 7. Tests de Stress (`stress-test.js`)
- Montée en charge progressive
- Test de pic de trafic
- Charge : 50-200 utilisateurs virtuels
- Seuils d'erreur plus élevés

## 📈 Métriques et Seuils

### Seuils Standard
- **Temps de réponse** : P95 < 1000ms, P99 < 2000ms
- **Taux d'erreur** : < 5%
- **Disponibilité** : > 99%

### Seuils de Stress
- **Temps de réponse** : P95 < 3000ms, P99 < 5000ms
- **Taux d'erreur** : < 15%

## 📁 Résultats

Les résultats sont automatiquement sauvegardés dans le dossier `results/` :

```
results/
├── ping-test-results-local.json
├── login-test-results-staging.json
├── stress-test-results-prod.json
└── ...
```

## 🔧 Configuration

### Variables d'Environnement

- `ENVIRONMENT` : `local`, `staging`, `prod`

### Paramètres de Test

Chaque test peut être configuré en modifiant les options dans le fichier :

```javascript
export const options = {
  scenarios: {
    // Configuration des scénarios
  },
  thresholds: {
    // Seuils de performance
  },
};
```

## 🎯 Scénarios de Test

### Développement
```powershell
.\tests\k6\run-tests.ps1 -TestType basic
```

### Avant Déploiement
```powershell
.\tests\k6\run-tests.ps1 -TestType all -Environment staging
```

### Validation Production
```powershell
.\tests\k6\run-tests.ps1 -TestType full-flow -Environment prod
```

### Test de Performance
```powershell
.\tests\k6\run-tests.ps1 -TestType stress -Environment prod
```

## 🚨 Dépannage

### Problèmes Courants

1. **Serveur non démarré** : Vérifiez que l'API est accessible
2. **Erreurs d'authentification** : Vérifiez les credentials dans les tests
3. **Timeouts** : Ajustez les seuils selon votre environnement

### Logs et Debug

Utilisez `--verbose` pour plus de détails :
```bash
k6 run --verbose ping.js
```

## 📚 Ressources

- [Documentation K6](https://k6.io/docs/)
- [Guide des Tests de Performance](https://k6.io/docs/testing-guides/)
- [Métriques K6](https://k6.io/docs/using-k6/metrics/)
