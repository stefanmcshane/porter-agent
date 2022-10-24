## Initial Setup

### Prereq steps if you are not running `porter-agent` on the same cluster that you are about to install ArgoCD on.

- Install [ngrok](https://ngrok.com)
- Assuming that `porter-agent` is listening on port 10001, run

`ngrok http 10001 --host-header="localhost:10001"`

- Copy the "forwarding" address into `values.yaml` at `.notifications.notifiers.service.webhook.agent.url`
- Ensure the address above ends in `/listen/argocd`

### Install ArgoCD

Install argo using the `values.yaml`

`helm install -n argocd --create-namespace -f values.yaml argo-cd argo/argo-cd`

### Getting Access via CLI as admin

Port forward the argocd-server to the port of your choice

`kubectl port-forward service/argo-cd-porter-server -n argocd 8080:443`

Read the admin password from kubernetes secrets (if it wasnt already set through `values.yaml`)

`export PW=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)`

Log into the argocd cluster via CLI (assuming that you exposed port 8080 in the above steps)

`argocd login localhost:8080 --insecure --username=admin --password=$PW`

### Create an application to be monitored by ArgoCD

Create a sample application

`kubectl apply -f sampleApplication.yaml`

### Generate API Token

The account name `porter` should already exist as it was created in the helm install.
Ensure that you are logged in as admin, and run the following:

`argocd account generate-token --account porter`

TODO: Add this as API function

### Test notification sending

- Watch the logs on your porter-agent
- Run `kubectl scale deploy nginx-test-web --replicas=3`
- ArgoCD will force this back to 2 replicas
- Check the porter-agent logs, you should see a notification that has come through for each sync state

## Read events

- Ensure porter-agent is running (on port 10001)
- Run `curl localhost:10001/events`

```
Received argo hook: types.ArgoCDResourceHook{Application:"guestbook", Status:"OutOfSync", Author:"Stefan McShane <stefanmcshane@users.noreply.github.com>", Timestamp:"2022-10-20 05:13:02.025725757 +0000 UTC m=+21.390802220"}
2022/10/20 01:13:02 "POST http://localhost:10001/listen/argocd HTTP/1.1" from [::1]:54727 - 000 0B in 42.958µs

Received argo hook: types.ArgoCDResourceHook{Application:"guestbook", Status:"Synced", Author:"Stefan McShane <stefanmcshane@users.noreply.github.com>", Timestamp:"2022-10-20 05:13:02.155096507 +0000 UTC m=+21.520172970"}
2022/10/20 01:13:02 "POST http://localhost:10001/listen/argocd HTTP/1.1" from [::1]:54727 - 000 0B in 143.833µs
```

## Notes

`values.yaml` was pulled from [argo-helm](https://github.com/argoproj/argo-helm) and the following sections have been added or updated:

- `.notifications.notifiers`
- `.notifications.templates`
- `.notifications.triggers`
- `.notifications.logging.format` = "json"
- `.notifications.logging.format` = "json"
- `.global.logging.format` = "json"
- `.global.logging.level` = "debug"
- removed `.dex` (unused and can be pulled in later if needs be)
- removed `.redis-ha` (unused, only using standard redis for now, can use externalRedis if needs be)
- removed `.externalRedis`
- removed `.applicationSet`

## Work to be done

- Disable admin secret being added to cluster through `values.yaml` a
- Set custom password for admin in `values.yaml` at `argocdServerAdminPassword`
- Setup TLS
- Setup ingress through Nginx if exposing outside cluster `.server.ingress`
- Enable promethus metrics
- enable redis-ha, or externalRedis (not needed, but will make argo faster if needs be)
- Set resource limits for notifications, repo-server, server, redis
- Setup autoscaling on server
- Increase default replicas on server
