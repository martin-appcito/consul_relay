package main

import (
	"fmt"
//	"strings"
	"errors"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type ConsulNode struct {
	Node    	string
	Address 	string
	ServiceId 	string
	ServiceName	string
	ServiceTags	string
	ServicePort	int
}

type AppcitoNode struct {
	AccountId      string
	PublicAddress  string
	PrivateAddress string
	InternalPort   int
	ExternalPort   int
}

type AwsIdentity struct {
	privateIp			string	
  	availabilityZone	string
	accountId			string
  	region				string
  }

func getServices() ([]string, error) {
	
	var f interface{}
	services := make([]string, 0)
	
	resp, err := http.Get("http://consul:8500/v1/catalog/services")
	if err != nil {
		return nil,err
	}
	
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil,err
	}
	
	err = json.Unmarshal(body, &f)
	if err != nil {
		return nil,err
	}
	
	m := f.(map[string]interface{})
	i := 0
	
	for k, v := range m {
		switch v.(type) {
		case []interface{}:
			services = append(services, k);
			i++
		default:
			return nil,errors.New(k + " - unexpected type");
		}
	}

	return services, nil
}
 
 func getAccountId() (AwsIdentity, error) {
 	var acct AwsIdentity
 	
 	resp, err := http.Get("http://169.254.169.254/latest/dynamic/instance-identity/document")
	if err != nil {
		return acct,err
	}
	
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return acct,err
	}
	
	err = json.Unmarshal(body, &acct)
	if err != nil {
		return acct,err
	}
	
	return acct, nil
}
func main() {

	var nodes []ConsulNode
	topology := make(map[string][]ConsulNode)
	var acct AwsIdentity

	
	services, err := getServices()
	if err != nil {
		fmt.Println(err)
		return
	}
	
	acct, err = getAccountId()
	if err != nil {
		fmt.Println(err)
		return
	}
	
	fmt.Println(json.MarshalIndent(acct, "", "    "))
	
	for i := range services {

		resp, err := http.Get("http://consul:8500/v1/catalog/service/" + 
			services[i])

		if err != nil {
			fmt.Print(resp.Status)
			return
		}
		
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		json.Unmarshal(body, &nodes)

		for i := range nodes {
			
			sname := nodes[i].ServiceName
			if topology[sname] == nil {
				topology[sname] = make([]ConsulNode, 0)
			}
			topology[sname] = append(topology[sname], nodes[i]);
		}
	}
	
	for s := range services {
	
		for n := range topology[services[s]] {
				
			fmt.Printf("%s\n", topology[services[s]][n].Node)
			
		}
	}
	
	j, err := json.Marshal(topology)
	
	fmt.Printf("%s", j)
}
