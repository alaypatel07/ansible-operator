## TESTING EXAMPLE ANSIBLE OPERATOR

1. Bring up [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
2. Clone this repo by 
    ```
        mkdir -p $GOPATH/src/github.com/water-hole && cd $GOPATH/src/github.com/water-hole && git clone https://github.com/alaypatel07/ansible-operator.git
    ```
3. Run 
    ```
    dep ensure -v -vendor-only
    ```
3. Watch the kubernete resources in all namespaces
    
    ```
     watch kubectl get all --all-namespaces 
     ```
     

4. Run the manual go tests 
    ```
    go test ./test/e2e/... -root=$(pwd) -kubeconfig=$HOME/.kube/config -globalMan example/deploy/crd.yaml -namespacedMan example/deploy/namespace-init.yaml -v -parallel=2
    ```

5. If you have latest version of [operator-sdk](https://github.com/operator-framework/operator-sdk), run the tests using operator sdk as follows: 
    ```
    operator-sdk test --test-location ./test/e2e -n $(pwd)/example/deploy/namespace-init.yaml -g $(pwd)/example/deploy/crd.yaml
    ```
