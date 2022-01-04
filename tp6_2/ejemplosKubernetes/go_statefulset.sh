kubectl delete statefulset raft
kubectl delete service elixir
kubectl delete configmap cm-elixir
kubectl create configmap cm-elixir --from-file=./prueba.yaml
kubectl create -f statefulset_go.yaml
