helm install argo/argo-cd argo-cd -f argovalues.yaml -n argocd --create-namespace

kubectl port-forward service/argo-cd-argocd-server -n argocd 8080:443

kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

argocd login localhost:8080

argocd app create guestbook --repo https://github.com/argoproj/argocd-example-apps.git --path guestbook --dest-namespace default --dest-server https://kubernetes.default.svc --directory-recurse

# run this after `k apply -f user.yaml`, but make sure admin isnt disabled

argocd account generate-token --account porter

wSPrTqH54gzg3Yt7
