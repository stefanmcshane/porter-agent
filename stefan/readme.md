Initial Install:

helm install -n argocd --create-namespace -f argovalues.yaml argo-cd argo/argo-cd

kubectl port-forward service/argo-cd-argocd-server -n argocd 8080:443

export PW=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)

argocd login localhost:8080 --insecure --username=admin --password=$PW

argocd app create guestbook --repo https://github.com/argoproj/argocd-example-apps.git --path guestbook --dest-namespace default --dest-server https://kubernetes.default.svc --directory-recurse

k apply -f permissions.yaml -f user.yaml

argocd account generate-token --account porter

Updates:
helm upgrade -n argocd -f argovalues.yaml argo-cd argo/argo-cd

Delete later:
gf4-KB9Smrue7irv

eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhcmdvY2QiLCJzdWIiOiJwb3J0ZXI6YXBpS2V5IiwibmJmIjoxNjY2MjI3MDU1LCJpYXQiOjE2NjYyMjcwNTUsImp0aSI6IjVmM2U0NGRkLWEyODktNDQyYi05NzE5LWYzY2MyNWEwOTVlZiJ9.73R8QPr6ps4vAtvfkhbeLM4YhkRSbHx-bOlFXP6L00M
