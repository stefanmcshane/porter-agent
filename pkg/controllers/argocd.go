package controllers

// import (
// 	"context"
// 	"fmt"

// 	argocd "github.com/argoproj/argo-cd/v2/pkg/apiclient"
// 	argoapp "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
// )

// type ArgoCD struct {
// 	Client argocd.Client
// }

// func NewArgoCDClient() (ArgoCD, error) {
// 	var cli ArgoCD

// 	opts := argocd.ClientOptions{
// 		ServerAddr: "localhost:8080",
// 		Insecure:   true,
// 		AuthToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhcmdvY2QiLCJzdWIiOiJwb3J0ZXI6YXBpS2V5IiwibmJmIjoxNjY2MTUxMDI3LCJpYXQiOjE2NjYxNTEwMjcsImp0aSI6ImZkODg3ZWZjLTE1N2UtNDdhMC05OTY3LTk3MjI2OWY1ODM1MSJ9.7_xx879dOde4Z82DTV2SnW5eNzEWAw1-mLjn0cltTG8",
// 	}
// 	client, err := argocd.NewClient(&opts)
// 	if err != nil {
// 		return cli, err
// 	}
// 	appClo, appCli, err := client.NewApplicationClient()
// 	if err != nil {
// 		return cli, err
// 	}
// 	defer appClo.Close()

// 	ctx := context.Background()
// 	applicationName := "test"
// 	in := argoapp.ApplicationQuery{
// 		Name: &applicationName,
// 	}
// 	application, err := appCli.Get(ctx, &in)
// 	if err != nil {
// 		return cli, err
// 	}

// 	fmt.Println("STEFAN", application)

// 	return cli, nil
// }

// func (a ArgoCD) Consume() {

// }
