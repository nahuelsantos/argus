# Argus Version Management

This document explains how version management works in Argus.

## 🎯 Single Source of Truth

**Version is managed from ONE place**: `internal/config/VERSION`

```
argus-repo/internal/config/VERSION
```

This file contains just the version number (e.g., `0.0.1`) without the `v` prefix.

## 🔄 How Versioning Works

### 1. **Build-time Injection (Preferred)**
Go build uses `-ldflags` to inject version at build time:

```bash
go build -ldflags "-X 'github.com/nahuelsantos/argus/internal/config.Version=v1.0.0'"
```

### 2. **Environment Variable Fallback**
If build-time version isn't set, checks `SERVICE_VERSION` env var:

```bash
export SERVICE_VERSION=v1.0.0
./argus
```

### 3. **VERSION File Fallback**
If no env var, reads from embedded `VERSION` file.

### 4. **Ultimate Fallback**
If all else fails: `v0.0.1-dev`

## 🛠️ Version Management Commands

### **Check Current Version**
```bash
make version
```

### **Update Version**
```bash
# Bump patch: 0.0.1 -> 0.0.2
make version-patch

# Bump minor: 0.1.0 -> 0.2.0  
make version-minor

# Bump major: 1.0.0 -> 2.0.0
make version-major

# Set specific version
make release VERSION=1.0.0
```

### **Build with Version**
```bash
# Build with automatic version from file
make build

# Build and run
make run

# Build Docker image with version
make docker-build
```

### **Sync Versions Across Files**
```bash
# Update Dockerfile, README, etc.
./scripts/update-version.sh
```

## 📦 Docker Building

### **Local Build**
```bash
# Uses VERSION file automatically
make docker-build
```

### **Manual Build with Version**
```bash
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  --build-arg GIT_COMMIT=$(git rev-parse HEAD) \
  -t argus:v1.0.0 .
```

### **GitHub Actions**
Automatically reads from `VERSION` file and builds with proper metadata.

## 🌐 Frontend Version Display

The frontend gets version from the API `/config` endpoint:

```javascript
// Version is fetched from API and displayed dynamically
fetch('/config').then(r => r.json()).then(config => {
    document.querySelector('.version').textContent = config.version;
});
```

## 🔄 Release Process

### **Manual Release**
```bash
# 1. Update version
echo "1.0.0" > internal/config/VERSION

# 2. Update all references
./scripts/update-version.sh

# 3. Build and test
make build
make run

# 4. Commit and tag
git add internal/config/VERSION
git commit -m "Bump version to v1.0.0"
git tag v1.0.0
git push origin v1.0.0
```

### **Automated Release**
```bash
# One command release
make release VERSION=1.0.0
```

## 🔍 Where Version Appears

### **Runtime**
- Go binary startup message
- API `/config` endpoint response
- Web dashboard header
- Docker container labels
- Log messages

### **Build Artifacts**
- Docker image tags
- GitHub release artifacts
- Binary metadata

### **Documentation**
- README.md examples
- Documentation references
- Docker run commands

## 🎛️ Integration with Dinky Server

Dinky Server references Argus version in multiple files. Use the sync script:

```bash
# From dinky-server root
./scripts/sync-argus-version.sh
```

This updates:
- `Makefile`
- `dinky.sh` 
- `README.md`
- `docs/apis-guide.md`

## ✅ Best Practices

1. **Always update VERSION file first**
2. **Use scripts to sync across files**
3. **Test build before releasing**
4. **Tag releases in git**
5. **Use semantic versioning**

## 🚨 Important Notes

- Never hardcode versions in Go code
- Always use `config.GetVersion()` function
- Docker images get version from build args
- Frontend gets version from API (dynamic)
- Keep VERSION file clean (just the number)

## 🔧 Troubleshooting

### **Version shows as v0.0.1-dev**
- Check if VERSION file exists
- Verify build used proper ldflags
- Check environment variables

### **Frontend shows wrong version**
- Clear browser cache
- Check API `/config` endpoint
- Verify service restart after version change

### **Docker build issues**
- Ensure VERSION file is in build context
- Check build args are passed correctly
- Verify Dockerfile ARG declarations 