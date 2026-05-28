# VS Code

Browser-based VS Code for the cluster, built from code-server. Editor settings
and extensions are expected to be mounted at runtime.

## Runtime

The image runs as `1000:1000` (`coder:coder`) by default and only uses the
paths you mount into the container:

- `/config` for code-server user data and installed extensions
- `/home/coder/src` for source checkouts
- `/opt/code-server-defaults/settings.json` for managed VS Code settings
- `/opt/code-server-defaults/extensions.txt` for managed extension IDs

Do not mount the host root or broad host paths into `/home/coder/src`. In
Kubernetes, prefer a PVC or a narrow repository checkout volume.

Set `PASSWORD` or `HASHED_PASSWORD` from a Secret unless you put the service
behind another authenticated proxy and set `CODE_SERVER_AUTH=none`.

Recommended pod security context:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000
```

Recommended container security context:

```yaml
securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  readOnlyRootFilesystem: false
```
