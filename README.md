# maskcmd

Wrapper for cli commands to mask unwanted output (e.g. secrets, credentials).
Useful for bash scripts within K8S native pipelines, like Argo Workflows.

Usage example:

```bash
# simulate argo workflow context with mounted secrets
mkdir -p /tmp/maskcmd-tmp
echo "password" > /tmp/maskcmd-tmp/db-password

# actual command
./maskcmd --secrets-dir /tmp/maskcmd-tmp -- bash -c "echo psql -W $(cat /tmp/maskcmd-tmp/db-password)"
# will produce: psql -W *****
```
