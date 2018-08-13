## Deploy from private Git repo

See [Basic Workflow](./basic-workflow.md) for introduction.

Create new namespace

```bash
$ knctl create namespace -n deploy-from-git

$ export KNCTL_NAMESPACE=deploy-from-git
```

Create SSH secret to pull Git repository

```bash
$ knctl create ssh-auth-secret -s git1 --github --private-key-path ~/.ssh/

# ... or for non-github.com urls ...

$ knctl create ssh-auth-secret -s git1 --type git --url gitlab.com --private-key-path ~/.ssh/
```

Create Docker Hub secret for pushing images

```bash
$ knctl create basic-auth-secret -s docker-reg1 --docker-hub -u <your-username> -p <your-password>
```

If necessary, create Docker Hub secret for pulling images

```bash
$ knctl create basic-auth-secret -s docker-reg2 --docker-hub -u <your-username> -p <your-password> --for-pulling
```

Create service account that references above credentials

```bash
$ knctl create service-account -a serv-acct1 -s git1 -s docker-reg1 [-s docker-reg2]
```

Deploy service that builds image from a Git repo, and then deploys it

```bash
$ knctl deploy \
    --service simple-app \
    --git-url git@github.com:cppforlife/simple-app-private \
    --git-revision master \
    --service-account serv-acct1 \
    --image index.docker.io/<your-username>/<your-repo> \
    --env SIMPLE_MSG=123
```
