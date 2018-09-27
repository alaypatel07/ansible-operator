package e2e

import (
	"testing"
	"time"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"io"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"context"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime"
	"errors"
	"path/filepath"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
)

var (
	retryInterval   = time.Second * 5
	timeout         = time.Second * 30
	testDir         = "./example/test"
	assertsFilePath = testDir + "/asserts.yaml"
	resourcesDir    = testDir + "/resources/"
)

type Assert struct {
	Resource map[string]string      `json:"resource"`
	Result   map[string]interface{} `json:"result"`
}

func TestExample(t *testing.T) {
	// run subtests
	t.Run("example-group", func(t *testing.T) {
		t.Run("Cluster", ExampleCluster)
	})
}

func readAsserts(reader io.Reader) ([]Assert, error) {
	d := yaml.NewYAMLOrJSONDecoder(reader, 65535)
	var asserts []Assert
	err := d.Decode(&asserts)
	return asserts, err
}

func readCrs(reader io.Reader) (*unstructured.Unstructured, error) {
	d := yaml.NewYAMLOrJSONDecoder(reader, 65535)
	var cr unstructured.Unstructured;
	err := d.Decode(&cr)
	return &cr, err
}

func ExampleCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Logf("could not get namespace: %v", err)
	}
	defer ctx.Cleanup(t)
	var crFiles []string

	err = filepath.Walk(resourcesDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			crFiles = append(crFiles, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	crs := make([]*unstructured.Unstructured, 0)

	for _, fp := range crFiles {
		f, err := os.Open(fp)
		if err != nil {
			panic(err)
		}
		c, err := readCrs(f)
		if err != nil {
			panic(err)
		}
		c.SetNamespace(namespace)
		crs = append(crs, c)
	}

	ctx.AddFinalizerFn(func() error {
		for _, cr := range crs {
			framework.Global.DynamicClient.Delete(context.TODO(), cr)
		}
		return nil
	})
	fmt.Println("start")
	err = ctx.InitializeClusterResources()
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	fmt.Println("Initialized")
	t.Log("Initialized cluster resources")
	fmt.Println("namespaced")
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for memcached-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "ansible-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Deployment successfull")

	fmt.Println("Registerging appGVK")

	appGVK := schema.GroupVersion{
		Group:   "app.example.com",
		Version: "v1alpha1",
	}


	framework.AddToFrameworkScheme(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(appGVK, &unstructured.Unstructured{},
			&unstructured.UnstructuredList{})
		metav1.AddToGroupVersion(scheme, appGVK)
		return nil
	}, &unstructured.UnstructuredList{
		Items: nil,
	})

	for _, cr := range crs {
		err = framework.Global.DynamicClient.Create(context.TODO(), cr)
		if err != nil {
			t.Logf("error creating crs %+v\n", err)
			t.Fail()
		}
	}

	r, err := os.Open(assertsFilePath)
	if err != nil {
		fmt.Printf("error reading asserts.yaml %s\n", err.Error())
		t.Logf("error reading asserts.yaml %s\n", err.Error())
	}

	asserts, err := readAsserts(r)
	if err != nil {
		fmt.Printf("error loading asserts %s\n", err.Error())
		t.Logf("error loading asserts %s\n", err.Error())
	}

	resources := make([]*unstructured.Unstructured, 0)
	results := make([]map[string]interface{}, 0)

	fmt.Printf("Printing asserts\n")

	for _, assert := range asserts {
		//fmt.Printf("%d, %+v\n", i, assert)
		u := unstructured.Unstructured{}
		gv, err := schema.ParseGroupVersion(assert.Resource["apiVersion"])
		if err != nil {
			panic("error converting gvk")
		}
		gvk := gv.WithKind(assert.Resource["kind"])
		u.SetGroupVersionKind(gvk)
		resources = append(resources, &u)
		results = append(results, assert.Result)
	}
	fmt.Printf("Waiting for asserts\n")

	err = WaitForResources(t, resources, results, namespace, retryInterval, timeout)
	if err != nil {
		t.Logf("error matching asserts %s\n", err)
		t.Fail()
	}
	//err = WaitForDeployment(t, f.KubeClient, namespace, "test-example-busybox", 1, retryInterval, timeout)

}

func WaitForResources(t *testing.T, resources []*unstructured.Unstructured, results []map[string]interface{}, namespace string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		for i, r := range resources {

			u := &unstructured.Unstructured{}

			fmt.Printf("GVK is %+v %v\n", r.GroupVersionKind(), i)
			lo := crclient.InNamespace(namespace)

			u.SetGroupVersionKind(r.GroupVersionKind())
			//u.SetNamespace(r.GetNamespace())
			//u.SetName(r.GetName())

			err := framework.Global.DynamicClient.List(context.TODO(), lo, u)
			if err != nil {
				fmt.Printf("err getting list %s\n", err.Error())
			}
			counter := 0
			err = u.EachListItem(func(object runtime.Object) error {
				counter++
				return nil
			})
			if err != nil {
				panic("error calling each item")
			}

			if rc, ok := results[i]["number"].(float64); ok {
				if int(rc) != counter {
					fmt.Printf("counter is not equal expected counter %v\n", counter)
					t.Logf("waiting for the counter to equal expected number")
					return false, nil
				} else {
					t.Logf("counter is equal to expected counter")
					return true, nil
				}
			} else {
				fmt.Printf("cannot convert result.number to float64")
				return false, errors.New("cannot convert result.number to float64")
			}
		}
		return false, nil
	})
	return err
}