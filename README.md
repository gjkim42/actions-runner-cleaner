# actions-runner-cleaner

actions-runner-cleaner cleans up the runners that have been offline for a long
time.

```sh
go install github.com/gjkim42/actions-runner-cleaner@latest

export GITHUB_USERNAME=MY_USERNAME
export GITHUB_SECRET=MY_SECRET
actions-runner-cleaner --help
```

## Deploy actinos-runner-cleaner to the kubernetes

```sh
kubectl create secret generic github-secret \
  --from-literal=GITHUB_USERNAME=${MY_USERNAME} \
  --from-literal=GITHUB_SECRET=${MY_SECRET}

ORG=MY_ORG REPOSITORY=MY_REPOSITORY envsubst < actions-runner-cleaner.yaml | kubectl create -f -
```
