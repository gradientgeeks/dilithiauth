# DilithiAuth SDK Publishing Guide

This guide covers how to publish the **DilithiAuth Client SDKs** across **npm**, **PyPI**, **NuGet**, **Go Module Proxy**, and **Maven Central**.

---

## 🚀 Automated Publishing via GitHub Actions (Recommended)

The project includes an automated GitHub Actions pipeline ([.github/workflows/sdk-ci-cd.yml](file:///home/uttam/projects/DilithiAuth/.github/workflows/sdk-ci-cd.yml)).

### How to Trigger an Automated Release Publish:
Simply push a Git release tag (e.g. `v1.0.0` or `sdk-v1.0.0`) to GitHub:

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions will automatically build, test, and publish all 5 SDKs to their respective registries using your repository secrets:
- `NPM_TOKEN`: npm registry auth token.
- `PYPI_API_TOKEN`: PyPI API token (`pypi-xxx`).
- `NUGET_API_KEY`: NuGet API key.

---

## 📦 Manual Publishing Steps

### 1. JavaScript / TypeScript SDK (`sdk/js`) — npm
```bash
cd sdk/js
npm install
npm run build
npm publish --access public
```

### 2. Python SDK (`sdk/python`) — PyPI
```bash
cd sdk/python
pip install build twine
python -m build
twine upload dist/*
```

### 3. Go SDK (`sdk/go`) — Go Modules
Go modules are published automatically by tagging the git repository:
```bash
git tag sdk/go/v1.0.0
git push origin sdk/go/v1.0.0
```
Callers can then run: `go get github.com/Uttam-Mahata/DilithiAuth/sdk/go@v1.0.0`.

### 4. .NET SDK (`sdk/dotnet`) — NuGet
```bash
cd sdk/dotnet/DilithiAuth.Client
dotnet pack -c Release
dotnet nuget push bin/Release/*.nupkg -k YOUR_NUGET_API_KEY -s https://api.nuget.org/v3/index.json
```

### 5. Java SDK (`sdk/java`) — Maven Central
```bash
cd sdk/java
mvn clean deploy -P release
```
