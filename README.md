# maskcmd

Wrapper for cli commands to mask unwanted output (e.g. secrets, credentials).
Useful for bash scripts within K8S native pipelines, like Argo Workflows.

## Usage examples

### Argo Workflow 

```bash
kubectl apply -f argo-workflow-example/secret.yaml
argo submit argo-workflow-example/example.yaml -w --log
Name:                maskcmd-example-mm2nj
Namespace:           default
ServiceAccount:      unset (will run with the default ServiceAccount)
Status:              Pending
Created:             Sun Mar 16 22:42:25 +0100 (now)
Progress:
maskcmd-example-mm2nj: + git clone https://x-token-auth:*****@bitbucket.org/project1/repo1.git
maskcmd-example-mm2nj: Cloning into 'repo1'...
maskcmd-example-mm2nj: remote: You may not have access to this repository or it no longer exists in this workspace. If you think this repository exists and you have access, make sure you are authenticated.
maskcmd-example-mm2nj: fatal: repository 'https://bitbucket.org/project1/repo1.git/' not found
maskcmd-example-mm2nj: Error: child command returned exit code: 128
maskcmd-example-mm2nj: time="2025-03-16T21:42:28.457Z" level=info msg="sub-process exited" argo=true error="<nil>"
maskcmd-example-mm2nj: Error: exit status 128
maskcmd-example-mm2nj Failed at 2025-03-16 22:42:35 +0100 CET
```

Notice that there is no `bitbucket-repo1-token` (secret value) in the output, but just asterisks (`*****`).

[Example K8S manifests]([argo-workflow-example/example.yaml), [secret]([argo-workflow-example/sectet.yaml).

### Shell scripts

Mask files content in certain dir:

```bash
# imagine Argo Workflow with mounted secrets
mkdir -p /tmp/maskcmd-tmp
echo "password" > /tmp/maskcmd-tmp/db-password

# actual command
./maskcmd --secrets-dir /tmp/maskcmd-tmp -- bash -c "echo psql -W $(cat /tmp/maskcmd-tmp/db-password)"
psql -W *****
```

Mask all environment variables values:

```bash
export SECRET=mysecret
./maskcmd --all-env-vars -- bash -c 'echo secret is $SECRET'
# probably number of "Warning: overlapping secrets detected..."
secret is *****
```

Mask only certain environment variables values: 

```bash
export SECRET=mysecret
./maskcmd --env-vars SECRET -- bash -c 'echo secret is $SECRET'
secret is *****
```

Original exit code is preserved:

```bash
export SECRET=mysecret
./maskcmd --env-vars SECRET -- sh -c "echo secret=mysecret; exit 5"
secret=*****
Error: child command returned exit code: 5
echo $?
5
```

## Docker image

[Dockerfile](/Dockerfile) is based on recent [alpine/git](https://hub.docker.com/r/alpine/git): https://hub.docker.com/r/caseycs/maskcmd/tags
