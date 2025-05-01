/*
(C) Copyright Hewlett Packard Enterprise Development LP
*/

// Package fm provides the functions to get the switches, and port details
package fm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.hpe.com/hpe/sshot-net-operator/httpclient"
	"github.hpe.com/hpe/sshot-net-operator/models"
)

// GetAllSwitches gets all the switches
func GetAllSwitches() ([]string, error) {
	ctx := context.Background()
	switches := []string{}
	httpClient := httpclient.NewClient(models.BaseURL)
	response, err := httpClient.SendRequest(ctx, "GET", "/fabric/switches", nil)
	if err != nil {
		return switches, fmt.Errorf("could not get all switches: %v", err)
	}

	var SwitchesResponse models.SwitchesResponse
	err = json.Unmarshal(response, &SwitchesResponse)
	if err != nil {
		return switches, fmt.Errorf("cannot unmarshal switches response: %v", err)
	}

	for _, x := range SwitchesResponse.DocumentLinks {
		splitDocumentLink := strings.Split(x, "/")
		switches = append(switches, splitDocumentLink[len(splitDocumentLink)-1])
	}

	return switches, nil
}

// GetSwitch gets the ports for a switch
func GetSwitch(switchName string) (models.DFAComponents, error) {
	ctx := context.Background()
	var DFAComponents models.DFAComponents
	httpClient := httpclient.NewClient(models.BaseURL)
	response, err := httpClient.SendRequest(ctx, "GET", models.OperatorConstFabric+models.OperatorConstSwitches+switchName, nil)
	if err != nil {
		return DFAComponents, fmt.Errorf("could not get switch details: %v", err)
	}

	var SwitchResponse models.SwitchResponse
	err = json.Unmarshal(response, &SwitchResponse)
	if err != nil {
		return DFAComponents, fmt.Errorf("cannot unmarshal switch response: %v", err)
	}

	DFAComponents.GroupID = SwitchResponse.GrpID
	DFAComponents.SwitchID = SwitchResponse.SwcNum

	var edgePortInfo models.EdgePortsInfo
	for _, x := range SwitchResponse.EdgePorts {
		edgePortInfo.PortID = x.PortNum
		edgePortInfo.EdgePort = x.ConnPort
		DFAComponents.EdgePortsInfo = append(DFAComponents.EdgePortsInfo, edgePortInfo)
	}

	return DFAComponents, nil
}

// GetPort gets the port details for a port
func GetPort(portName string) (models.PortResponse, error) {
	ctx := context.Background()
	var port models.PortResponse

	httpClient := httpclient.NewClient(models.BaseURL)
	response, err := httpClient.SendRequest(ctx, "GET", models.OperatorConstFabric+models.OperatorConstPorts+portName, nil)
	if err != nil {
		return port, fmt.Errorf("could not get port details %+v", err)
	}

	err = json.Unmarshal(response, &port)
	if err != nil {
		return port, fmt.Errorf("cannot unmarshal port response: %+v", err)
	}

	return port, nil
}
