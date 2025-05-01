/*
(C) Copyright Hewlett Packard Enterprise Development LP
*/

// Package tapms defines controller for TAPMS resource
package tapms

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.hpe.com/hpe/sshot-net-operator/fm"
	"github.hpe.com/hpe/sshot-net-operator/httpclient"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	slingshot "github.hpe.com/hpe/sshot-net-operator/api/slingshot/v1alpha1"
	tapms "github.hpe.com/hpe/sshot-net-operator/api/tapms/v1alpha2"
	"github.hpe.com/hpe/sshot-net-operator/models"
	core "k8s.io/api/core/v1"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// VLANIDs is a global variable to store existing VLAN IDs
var (
	VLANIDs             = [256]int{}
	tenantList          tapms.TenantList
	slingshotTenantList slingshot.SlingshotTenantList
	VNIPartitionsList   models.AllVNIPartitionsResponse
	VNIBlocksList       models.AllVNIBlocksResponse
	httpClient          = httpclient.NewClient(models.BaseURL)
	tenantsMap          = make(map[string]tenantInfo)
	ClientID            = "admin-client"
)

type tenantInfo struct {
	tenantName       string
	tenantGeneration int64
	tenantXnames     []string
}

//+kubebuilder:rbac:groups=tapms.hpe.com.hpe.com,resources=tenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tapms.hpe.com.hpe.com,resources=tenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tapms.hpe.com.hpe.com,resources=tenants/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tenant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	accessToken, err := r.GetAccessToken(ctx)
	if err != nil {
		log.Printf("cannot get access token: %+v", err)
		return ctrl.Result{}, err
	}
	models.AccessToken = accessToken

	// Get the list of all v1alpha1 tenants
	var tl tapms.TenantList
	if err := r.List(ctx, &tl); err != nil {
		log.Printf("cannot list tenants: %+v", err)
		return ctrl.Result{}, nil
	}
	tenantList = tl

	//Get list of slingshot tenants
	var sTL slingshot.SlingshotTenantList
	if err := r.List(ctx, &sTL); err != nil {
		log.Printf("cannot list slingshot tenants: %s", err)
		return ctrl.Result{}, err
	}
	slingshotTenantList = sTL

	// Get all the VNI Partitions
	vniPartitions, err := GetAllVNIPartitions()
	if err != nil {
		log.Printf("cannot get VNI partitions: %s", err)
		return ctrl.Result{}, err
	}
	VNIPartitionsList = vniPartitions

	// Get all the VNI Blocks
	vniBlocks, err := GetAllVNIBlocks()
	if err != nil {
		log.Printf("cannot get VNI blocks: %s", err)
		return ctrl.Result{}, err
	}
	VNIBlocksList = vniBlocks

	if len(tenantsMap) == 0 && len(tenantList.Items) > 0 {
		var tt tenantInfo
		for _, tenant := range tenantList.Items {
			for _, t := range tenant.Spec.TenantResources {
				tt = tenantInfo{
					tenantName:       tenant.Spec.TenantName,
					tenantGeneration: tenant.Generation,
					tenantXnames:     t.XNames,
				}
			}
			tenantsMap[tenant.Name] = tt
		}
	}

	if len(tenantList.Items) > 0 {
		log.Println("checking if VNI partitions and VLAN are present for tenants")
		//Check for tenant creation. Compare the tenants and VNI Partitions
		for _, tenant := range tenantList.Items {
			// Check if tenant is present in VNI Partitions
			log.Printf("checking tenant %s", tenant.Spec.TenantName)
			var vniPartitionFound bool
			var vlanFound bool
			var vniBlockFound bool
			if len(vniPartitions.DocumentLinks) > 0 {
				for _, vniPartition := range vniPartitions.DocumentLinks {
					v := strings.Split(vniPartition, "/")
					if tenant.Spec.TenantName == v[len(v)-1] {
						vniPartitionFound = true
						log.Println("VNI partition exists:", vniPartition)
						break
					}
				}
			}

			//Iterate over slingshot tenants to get the tenant with same name as tenant
			var sshotTenant slingshot.SlingshotTenant
			var sshotTenantFound bool
			for _, sshotTenant = range slingshotTenantList.Items {
				if tenant.Spec.TenantName == sshotTenant.Spec.TenantName {
					sshotTenantFound = true
					log.Println("slingshot tenant exists:", sshotTenant.Spec.TenantName)
					for _, vniBlock := range VNIBlocksList.DocumentLinks {
						v := strings.Split(vniBlock, "/")
						if fmt.Sprintf("%s-%s", tenant.Spec.TenantName, sshotTenant.Spec.VNIBlockName) == v[len(v)-1] {
							vniBlockFound = true
							log.Println("VNI block exists:", vniBlock)
							break
						}
					}
					break
				}
			}

			if !sshotTenantFound {
				log.Println("cannot find slingshot tenant for tenant:", tenant.Spec.TenantName)
				continue
			}

			if !vniPartitionFound && sshotTenantFound {
				var txnames []string
				// Create the VNI Partition
				err := HandleCreate(ctx, &tenant, sshotTenant)
				if err != nil {
					log.Printf("cannot create VNI partition: %s", err)
					return ctrl.Result{}, err
				}
				tenantsMap[tenant.Name] = tenantInfo{
					tenantName:       tenant.Spec.TenantName,
					tenantGeneration: tenant.Generation,
				}
				for _, t := range tenant.Spec.TenantResources {
					txnames = append(txnames, t.XNames...)
				}
				info := tenantsMap[tenant.Name]
				info.tenantXnames = txnames
				tenantsMap[tenant.Name] = info
			}

			//Check if VLAN exists for the tenant
			vlanFound, _, err = CheckVLANExists(&tenant)
			if err != nil {
				log.Printf("cannot check if VLAN exists: %+v", err)
				return ctrl.Result{}, err
			}

			//if VLAN does not exist, create VLAN
			if !vlanFound {
				// Create VLAN
				var tenantXnames []string
				for _, t := range tenant.Spec.TenantResources {
					tenantXnames = append(tenantXnames, t.XNames...)
				}

				_, edgePorts, err := GetEdgePortDFAList(tenantXnames)
				if err != nil {
					log.Printf("cannot get edge ports for tenant: %+v", err)
					return ctrl.Result{}, err
				}

				vlan, err := CreateVLAN(edgePorts, tenant.Spec.TenantName)
				if err != nil {
					log.Printf("cannot create VLAN for tenant: %+v", err)
					return ctrl.Result{}, err
				}
				log.Printf("created VLAN %s for the tenant %s", vlan, tenant.Spec.TenantName)
			}

			//Check if VNI block exists for the tenant. If not, create VNI block
			if !vniBlockFound && sshotTenantFound {
				//create VNI block
				vniBlock, err := CreateVNIBlock(ctx, tenant, sshotTenant)
				if err != nil {
					log.Printf("cannot create VNI block: %+v", err)
					return ctrl.Result{}, err
				}
				log.Printf("created VNI block %s for the tenant %s", vniBlock.DocumentSelfLink, tenant.Spec.TenantName)

				//Check the stage of VniBlockEnforceTaskServiceState, keep checking until it is "FINISHED" or "FAILED"
				stage, err := CheckVniBlockEnforceTaskServiceState(ctx, vniBlock.EnforcementTaskServiceLink)
				if err != nil {
					log.Printf("cannot check VniBlockEnforceTaskServiceState: %+v", err)
					return ctrl.Result{}, err
				}

				if stage {
					log.Printf("enforcement for VNI block %s is completed", vniBlock.DocumentSelfLink)
				} else {
					log.Printf("enforcement for VNI block %s is failed", vniBlock.DocumentSelfLink)
				}
			}

			//Check if both tenant specification and slingshot tenant specification exists.
			//if yes, check if the generation of the tenant has changed. If yes, update the VNI partition.
			//if tenant Xname is updated, delete the previous VLAN and create a new VLAN. This
			//will be handled in the update function
			if vniPartitionFound && sshotTenantFound {
				if tenantsMap[tenant.Name].tenantGeneration != tenant.Generation {
					// Update the VNI Partition
					err := HandleUpdate(ctx, &tenant, sshotTenant)
					if err != nil {
						log.Printf("cannot update VNI partition or block: %+v", err)
						return ctrl.Result{}, err
					}
					tempTenantInfo := tenantsMap[tenant.Name]
					tempTenantInfo.tenantName = tenant.Spec.TenantName
					tempTenantInfo.tenantGeneration = tenant.Generation
					var txn []string
					for _, t := range tenant.Spec.TenantResources {
						txn = append(txn, t.XNames...)
					}
					tempTenantInfo.tenantXnames = txn
					tenantsMap[tenant.Name] = tempTenantInfo
				}
			}
		}
	}

	//check for tenant deletion. Check if tenant is marked for deletion and delete the VNI partition
	if len(tenantsMap) > 0 {
		tenantName := tenantsMap[req.Name].tenantName
		var tenantForDeletion tapms.Tenant
		err = r.Get(ctx, req.NamespacedName, &tenantForDeletion)
		if err != nil {
			var vniBlockName string
			for _, slingshotTenants := range slingshotTenantList.Items {
				if slingshotTenants.Spec.TenantName == tenantName {
					vniBlockName = fmt.Sprintf("%s-%s", tenantName, slingshotTenants.Spec.VNIBlockName)
					break
				}
			}
			if tenantName != "" {
				log.Printf("tenant %s is deleted. deleting VNI block %s, partition %s and VLAN", tenantName, vniBlockName, tenantName)
				err := HandleDelete(ctx, tenantName, vniBlockName)
				if err != nil {
					log.Printf("cannot delete VNI partition: %+v", err)
					return ctrl.Result{}, err
				}

				//delete the vlan for the tenant
				vlans, err := GetVLANs()
				if err != nil {
					log.Printf("cannot get all VLANs: %+v", err)
					return ctrl.Result{}, err
				}

				for _, vlan := range vlans {
					vlanID, err := strconv.Atoi(vlan[len(vlan)-1:])
					if err != nil {
						log.Printf("cannot convert vlan ID to integer: %+v", err)
						return ctrl.Result{}, err
					}
					vlanDetails, err := GetVLAN(vlanID)
					if err != nil {
						log.Printf("cannot get VLAN details: %+v", err)
						return ctrl.Result{}, err
					}
					if vlanDetails.VLANName == tenantName {
						err = DeleteVLAN(tenantName, strconv.Itoa(vlanID))
						if err != nil {
							log.Printf("cannot delete VLAN: %+v", err)
							return ctrl.Result{}, err
						}

					}
				}

				delete(tenantsMap, req.Name)
			}
		}
	}

	return ctrl.Result{RequeueAfter: models.ReconciliationTime}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tapms.Tenant{}).
		Complete(r)

}

// HandleCreate handles create events for tenant resource
func HandleCreate(ctx context.Context, tenant *tapms.Tenant, sshotTenant slingshot.SlingshotTenant) error {
	var vniRequestData models.VNIRequestData
	vniRequestData.PartitionName = tenant.Spec.TenantName
	vniRequestData.VNICount = sshotTenant.Spec.VNIPartition.VNICount
	vniRequestData.VNIRange = sshotTenant.Spec.VNIPartition.VNIRange

	// Validate the VNI request data
	err := ValidateVNIRequestData(vniRequestData)
	if err != nil {
		return err
	}

	var tenantXnames []string
	for _, t := range tenant.Spec.TenantResources {
		tenantXnames = append(tenantXnames, t.XNames...)
	}

	edgePortDFAList, _, err := GetEdgePortDFAList(tenantXnames)
	if err != nil {
		log.Printf("cannot get edge ports for tenant: %+v", err)
		return err
	}
	vniRequestData.EdgePortDFA = edgePortDFAList

	// Send the request
	responseBody, err := httpClient.SendRequest(ctx, "POST", "/fabric/vni/partitions", vniRequestData)
	if err != nil {
		log.Printf("cannot create VNI partition: %s %+v", responseBody, err)
		return err
	}

	//unmarshal the response in models.VNIPartitionResponse
	var vniPartition models.VNIPartitionResponse
	err = json.Unmarshal(responseBody, &vniPartition)
	if err != nil {
		log.Printf("cannot unmarshal VNI partition: %s %+v", responseBody, err)
		return err
	}

	log.Printf("created VNI partition %s for the tenant %s", vniPartition.DocumentSelfLink, sshotTenant.Spec.TenantName)

	return nil
}

// HandleUpdate handles create events for tenant resource
func HandleUpdate(ctx context.Context, tenant *tapms.Tenant, sshotTenant slingshot.SlingshotTenant) error {
	//check if tenant xname is updated. If yes, delete the previous VLAN and create a new VLAN
	var tenantXnameUpdated bool
	var tenantNodesCount int

	for _, t := range tenant.Spec.TenantResources {
		tenantNodesCount += len(t.XNames)
	}

	if tenantNodesCount != len(tenantsMap[tenant.Name].tenantXnames) {
		tenantXnameUpdated = true
	}

	if tenantNodesCount == len(tenantsMap[tenant.Name].tenantXnames) {
		for _, tr := range tenant.Spec.TenantResources {
			for _, xname := range tr.XNames {
				tenantXnameUpdated = true
				for _, xn := range tenantsMap[tenant.Name].tenantXnames {
					if xname == xn {
						tenantXnameUpdated = false
					}
				}
				if tenantXnameUpdated {
					break
				}
			}
		}
	}

	if tenantXnameUpdated {
		//check if VNI block name is empty
		if sshotTenant.Spec.VNIBlockName == "" {
			return fmt.Errorf("cannot update VNI block. VNIBlockName is empty")
		}

		var vniRequestData models.VNIRequestData
		vniRequestData.PartitionName = tenant.Spec.TenantName
		vniRequestData.VNICount = sshotTenant.Spec.VNIPartition.VNICount
		vniRequestData.VNIRange = sshotTenant.Spec.VNIPartition.VNIRange

		// Validate the VNI request data
		err := ValidateVNIRequestData(vniRequestData)
		if err != nil {
			return err
		}

		var tenantXnames []string
		for _, t := range tenant.Spec.TenantResources {
			tenantXnames = append(tenantXnames, t.XNames...)
		}

		edgePortDFAList, edgePorts, err := GetEdgePortDFAList(tenantXnames)
		if err != nil {
			log.Printf("cannot get edge ports for tenant: %+v", err)
			return err
		}
		vniRequestData.EdgePortDFA = edgePortDFAList

		// Send the request
		responseBody, err := httpClient.SendRequest(ctx, "PATCH", "/fabric/vni/partitions/"+tenant.Spec.TenantName, vniRequestData)
		if err != nil {
			log.Printf("cannot update VNI partition: %s %+v", responseBody, err)

			err = HandleDelete(ctx, tenant.Spec.TenantName, fmt.Sprintf("%s-%s", tenant.Spec.TenantName, sshotTenant.Spec.VNIBlockName))
			if err != nil {
				log.Printf("cannot update VNI partition: %+v", err)
				return err
			}

			_, vid, err := CheckVLANExists(tenant)
			if err != nil {
				log.Printf("cannot update VNI partition: %+v", err)
				return err
			}

			//delete the vlan for the tenant
			err = DeleteVLAN(tenant.Spec.TenantName, strconv.Itoa(vid))
			if err != nil {
				log.Printf("cannot delete VLAN: %+v", err)
				return err
			}

			//if partition not found, create the partition
			err = HandleCreate(ctx, tenant, sshotTenant)
			if err != nil {
				log.Printf("cannot create VNI partition: %+v", err)
				return err
			}
		}
		log.Println("updated VNI partitions for the tenant:", string(responseBody))

		//update VNI Block
		var vniBlockPatchRequestData models.VNIBlockPatchRequest
		vniBlockPatchRequestData.VNIBlockRange = sshotTenant.Spec.VNIPartition.VNIRange
		vniBlockPatchRequestData.PortDFAs = edgePortDFAList
		vniBlockName := fmt.Sprintf("%s-%s", tenant.Spec.TenantName, sshotTenant.Spec.VNIBlockName)

		//send the request
		vniBlockResponseBody, err := httpClient.SendRequest(ctx, "PATCH", "/fabric/vni/blocks/"+vniBlockName, vniBlockPatchRequestData)
		if err != nil {
			log.Printf("cannot update VNI block: %s %+v", vniBlockResponseBody, err)
			return err
		}

		var vniBlock models.VNIBlockResponse
		err = json.Unmarshal(vniBlockResponseBody, &vniBlock)
		if err != nil {
			log.Printf("cannot unmarshal VNI block: %s %+v", vniBlockResponseBody, err)
			return err
		}

		//Check the stage of VniBlockEnforceTaskServiceState, keep checking until it is "FINISHED" or "FAILED"
		stage, err := CheckVniBlockEnforceTaskServiceState(ctx, vniBlock.EnforcementTaskServiceLink)
		if err != nil {
			log.Printf("cannot check VniBlockEnforceTaskServiceState: %+v", err)
			return err
		}

		if stage {
			log.Printf("enforcement for VNI block %s is completed", vniBlock.DocumentSelfLink)
		} else {
			log.Printf("enforcement for VNI block %s is failed", vniBlock.DocumentSelfLink)
		}

		log.Println("updated VNI block for the tenant:", tenant.Spec.TenantName)

		log.Println("tenant xname is updated.updating vlan for the tenant:", tenant.Spec.TenantName)
		vlnaID, err := GetVlanID(ctx, tenant.Spec.TenantName)
		if err != nil {
			log.Printf("cannot get VLAN ID: %+v", err)
			return err
		}
		log.Println("deleting the previous vlan for the tenant:", tenant.Spec.TenantName)
		err = DeleteVLAN(tenant.Spec.TenantName, strconv.Itoa(vlnaID))
		if err != nil {
			log.Printf("cannot delete VLAN: %+v", err)
			return err
		}
		log.Println("creating new vlan for the tenant:", tenant.Spec.TenantName)
		vlan, err := CreateVLAN(edgePorts, tenant.Spec.TenantName)
		if err != nil {
			log.Printf("cannot create VLAN: %+v", err)
			return err
		}
		log.Printf("updated VLAN %s for the tenant %s", vlan, tenant.Spec.TenantName)

	}

	return nil
}

// HandleDelete handles create events for tenant resource
func HandleDelete(ctx context.Context, tenantName string, vniBlockName string) error {
	if tenantName == "" {
		log.Println("cannot delete vni partition. tenant name is empty")
		return nil
	}

	if vniBlockName == "" {
		log.Println("cannot delete vni block. vni block name is empty")
		return nil
	}

	log.Printf("deleting VNI enforced block %s for the tenant %s", vniBlockName, tenantName)
	//delete the VNI block
	_, err := httpClient.SendRequest(ctx, "DELETE", "/fabric/vni/blocks/"+vniBlockName, nil)
	if err != nil {
		log.Printf("cannot delete VNI block %s for the tenant:%s. %+v", vniBlockName, tenantName, err)
		return err
	}
	log.Printf("deleted VNI block %s", vniBlockName)

	// delete the VNI partition
	log.Printf("deleting VNI partition %s for the tenant %s", tenantName, tenantName)
	_, err = httpClient.SendRequest(ctx, "DELETE", "/fabric/vni/partitions/"+tenantName, nil)
	if err != nil {
		log.Printf("cannot delete VNI partition for the tenant:%s. %+v", tenantName, err)
		return err
	}

	log.Println("deleted VNI partition")

	return nil

}

// GetAccessToken gets the access token for CSM API gateway
func (r *TenantReconciler) GetAccessToken(ctx context.Context) (string, error) {
	//get the client-secret from admin-client-auth secret in the default namespace
	var adminSecret core.Secret
	var accessToken string

	err := r.Get(ctx, client.ObjectKey{Namespace: models.NamespaceForClientData, Name: models.SecretForClientData}, &adminSecret)
	if err != nil {
		log.Printf("cannot get admin-client-auth secret: %+v", err)
		return accessToken, err
	}

	adminClientSecretData := adminSecret.Data["client-secret"]
	slingshotAdminClientSecret := strings.TrimSpace(string(adminClientSecretData))

	//get endpoint for access token
	endpointBytes := adminSecret.Data["endpoint"]
	endpoint := string(endpointBytes)

	//create form data
	data := make(map[string]string)
	data["grant_type"] = "client_credentials"
	data["client_id"] = models.ClientID
	data["scope"] = "openid"
	data["client_secret"] = slingshotAdminClientSecret

	client := httpclient.NewClient("")

	resp, err := client.SendRequest(ctx, "POST", endpoint, data)
	if err != nil {
		log.Printf("cannot get access token: %+v", err)
		return accessToken, err
	}

	// parse the access token from the response body
	var result models.TokenResponse
	err = json.Unmarshal([]byte(resp), &result)
	if err != nil {
		log.Printf("cannot unmarshal access token: %+v", err)
		return accessToken, err
	}

	// get the access token
	accessToken = result.AccessToken

	return accessToken, nil

}

// GetPartition gets the VNI partition
func GetPartition(ctx context.Context, partitionName string) (models.VNIPartitionResponse, error) {
	responseBody, err := httpClient.SendRequest(ctx, "GET", "/fabric/vni/partitions/"+partitionName, models.VNIRequestData{})
	if err != nil {
		log.Printf("cannot get VNI partition: %s %+v", responseBody, err)
		return models.VNIPartitionResponse{}, err
	}

	//unmarshal the response in models.VNIPartitionResponse
	var vniPartition models.VNIPartitionResponse
	err = json.Unmarshal(responseBody, &vniPartition)
	if err != nil {
		log.Printf("cannot unmarshal VNI partition: %s %+v", responseBody, err)
		return models.VNIPartitionResponse{}, err
	}

	return vniPartition, nil
}

// GetEdgePortDFAList gets the list of edge ports for a tenant
func GetEdgePortDFAList(tenantXnames []string) ([]int, []string, error) {
	var edgePortDFAs []int
	var edgePorts []string

	switches, err := fm.GetAllSwitches()
	if err != nil {
		return edgePortDFAs, edgePorts, fmt.Errorf("could not get all switches %+v", err)
	}

	for _, x := range switches {
		DFAComponents, err := fm.GetSwitch(x)
		if err != nil {
			return edgePortDFAs, edgePorts, fmt.Errorf("could not get ports for switch %+v", err)
		}

		for _, p := range DFAComponents.EdgePortsInfo {
			port, err := fm.GetPort(p.EdgePort)
			if err != nil {
				return edgePortDFAs, edgePorts, fmt.Errorf("could not get port details for port %+v", err)
			}

			dstPort := port.DstPort

			for _, xname := range tenantXnames {
				if dstPort[:len(dstPort)-2] == xname {
					log.Printf("edge port found %s for xname %s", p.EdgePort, xname)
					edgePortDFA, err := CalculateEdgePortDFA(DFAComponents.GroupID, DFAComponents.SwitchID, p.PortID)
					if err != nil {
						return edgePortDFAs, edgePorts, fmt.Errorf("could not calculate edge port DFA %+v", err)
					}
					edgePortDFAs = append(edgePortDFAs, edgePortDFA)
					edgePorts = append(edgePorts, p.EdgePort)
				}
			}
		}

	}

	return edgePortDFAs, edgePorts, nil
}

// CalculateEdgePortDFA calculates the edge port DFA
func CalculateEdgePortDFA(grpID int, swID int, portID int) (int, error) {
	dfa := ((grpID << 23) | (swID << 18) | (portID << 12))
	return dfa, nil
}

// GetNewVLANID returns a new VLAN ID
func GetNewVLANID() (int, error) {
	err := GetExistingVLANIDs()
	if err != nil {
		log.Printf("cannot get existing VLAN IDs: %+v", err)
		return 0, err
	}

	var vlanid int

	for i := 0; i < len(VLANIDs); i++ {
		if VLANIDs[i] == 0 {
			vlanid = i + 1
			break
		}
	}

	return vlanid, nil
}

func createVlan(vlanid int, tenantname string) (string, error) {
	var vlanRequestData models.VLANRequestData
	vlanRequestData.VLANName = tenantname
	vlanRequestData.VLANID = vlanid
	vlanRequestData.Status = "ONLINE"

	httpClient := httpclient.NewClient(models.BaseURL)

	// send POST request to /fabric/vlans to create VLAN
	responseBody, err := httpClient.SendRequest(context.Background(), "POST", "/fabric/vlans", vlanRequestData)
	if err != nil {
		log.Printf("cannot create VLAN: %s %+v", responseBody, err)
		return "", err
	}

	var vlanResponse models.VLANResponse
	err = json.Unmarshal(responseBody, &vlanResponse)
	if err != nil {
		log.Printf("cannot unmarshal VLAN response: %+v", err)
		return "", err
	}

	return vlanResponse.DocumentSelfLink, nil
}

// CreateVLANPortPolicy creates VLAN port policy for a tenant
func CreateVLANPortPolicy(vlanid int, tenantname string) (models.VLANPortPolicyResponse, error) {
	var VLANPortPolicyResponse models.VLANPortPolicyResponse
	var VLANPortPolicyRequest models.VLANPortPolicyRequest
	VLANPortPolicyRequest.NativeVlanID = fmt.Sprintf("/fabric/vlans/%d", vlanid)
	VLANPortPolicyRequest.IsUntaggedAllowed = true
	VLANPortPolicyRequest.AllowedVlans = append(VLANPortPolicyRequest.AllowedVlans, fmt.Sprintf("/fabric/vlans/%d", vlanid))
	VLANPortPolicyRequest.DocumentSelfLink = tenantname

	httpClient := httpclient.NewClient(models.BaseURL)

	// send POST request to /fabric/port-policies to create VLAN port policy
	responseBody, err := httpClient.SendRequest(context.Background(), "POST", "/fabric/port-policies", VLANPortPolicyRequest)
	if err != nil {
		log.Printf("cannot create VLAN port policy: %s %+v", responseBody, err)
		return VLANPortPolicyResponse, err
	}

	err = json.Unmarshal(responseBody, &VLANPortPolicyResponse)
	if err != nil {
		log.Printf("cannot unmarshal VLAN port policy response: %+v", err)
		return VLANPortPolicyResponse, err
	}

	log.Printf("created VLAN port policy %s for tenant %s", VLANPortPolicyResponse.DocumentSelfLink, tenantname)
	return VLANPortPolicyResponse, nil
}

// ApplyVLANPortPolicyToEdgePorts applies VLAN port policy to edge ports
func ApplyVLANPortPolicyToEdgePorts(edgePorts []string, vlanPortPolicy models.VLANPortPolicyResponse) error {
	for _, edgePort := range edgePorts {
		log.Printf("applying VLAN port policy to edge port %+v", edgePort)
		var PortPATCHRequest models.PortPATCHRequest
		port, err := fm.GetPort(edgePort)
		if err != nil {
			log.Printf("cannot get port details for edge port: %+v", err)
			return err
		}

		for _, x := range port.PortPolicyLinks {
			if x == vlanPortPolicy.DocumentSelfLink {
				log.Printf("VLAN port policy already applied to edge port %s", edgePort)
				break
			}
		}

		PortPATCHRequest.PortPolicyLinks = append(PortPATCHRequest.PortPolicyLinks, vlanPortPolicy.DocumentSelfLink)
		PortPATCHRequest.PortPolicyLinks = append(PortPATCHRequest.PortPolicyLinks, port.PortPolicyLinks...)

		// send PATCH request to /fabric/ports/{edgePort} to apply VLAN port policy
		responseBody, err := httpClient.SendRequest(context.Background(), "PATCH", "/fabric/ports/"+edgePort, PortPATCHRequest)
		if err != nil {
			log.Printf("cannot apply VLAN port policy to edge port: %s %+v", responseBody, err)
			return err
		}
	}

	log.Printf("applied VLAN port policy to edge ports %+v", edgePorts)
	return nil
}

// CreateVLAN creates VLAN for a tenant
func CreateVLAN(edgePorts []string, tenantName string) (string, error) {
	log.Printf("creating VLAN for tenant %s", tenantName)

	vlanid, err := GetNewVLANID()
	if err != nil {
		log.Printf("cannot get new VLAN ID: %+v", err)
		return "", err
	}

	vlan, err := createVlan(vlanid, tenantName)
	if err != nil {
		log.Printf("cannot create VLAN: %+v", err)
		return "", err
	}

	//create VLAN port policy
	vlanPortPolicy, err := CreateVLANPortPolicy(vlanid, tenantName)
	if err != nil {
		log.Printf("cannot create VLAN port policy: %+v", err)
		return "", err
	}

	//apply vlan port policy to edge ports
	err = ApplyVLANPortPolicyToEdgePorts(edgePorts, vlanPortPolicy)
	if err != nil {
		log.Printf("cannot apply VLAN port policy to edge ports: %+v", err)
		return "", err
	}

	return vlan, nil
}

// GetVLANs gets the list of existing VLANs
func GetVLANs() ([]string, error) {
	ctx := context.Background()
	var VLANs []string

	responseBody, err := httpClient.SendRequest(ctx, "GET", "/fabric/vlans", nil)
	if err != nil {
		log.Printf("cannot get VLANs: %s %+v", responseBody, err)
		return VLANs, err
	}

	var VLANsResponse models.VLANsResponse
	err = json.Unmarshal(responseBody, &VLANsResponse)
	if err != nil {
		log.Printf("cannot unmarshal VLANs response: %+v", err)
		return VLANs, err
	}

	VLANs = append(VLANs, VLANsResponse.DocumentLinks...)

	return VLANs, nil
}

// GetVLAN gets the VLAN
func GetVLAN(vlan int) (models.VLANResponse, error) {
	ctx := context.Background()
	vlanLink := fmt.Sprintf("/fabric/vlans/%d", vlan)
	responseBody, err := httpClient.SendRequest(ctx, "GET", vlanLink, nil)
	if err != nil {
		log.Printf("cannot get VLAN: %s %+v", responseBody, err)
		return models.VLANResponse{}, err
	}

	var vlanResponse models.VLANResponse
	err = json.Unmarshal(responseBody, &vlanResponse)
	if err != nil {
		log.Printf("cannot unmarshal VLAN response: %s %+v", responseBody, err)
		return models.VLANResponse{}, err
	}

	return vlanResponse, nil
}

// GetExistingVLANIDs gets the list of existing VLAN IDs
func GetExistingVLANIDs() error {
	valns, err := GetVLANs()
	if err != nil {
		log.Printf("cannot get VLANs: %+v", err)
		return nil
	}

	for _, x := range valns {
		splitDocumentLink := strings.Split(x, "/")
		vlanID, err := strconv.Atoi(splitDocumentLink[len(splitDocumentLink)-1])
		if err != nil {
			log.Printf("cannot convert VLAN ID to integer: %+v", err)
			return err
		}
		VLANIDs[vlanID-1] = 1

	}
	return nil
}

// GetPortPolicy gets the port policy
func GetPortPolicy(portPolicy string) (models.PortPolicyResponse, error) {
	ctx := context.Background()
	httpClient := httpclient.NewClient(models.BaseURL)
	responseBody, err := httpClient.SendRequest(ctx, "GET", portPolicy, nil)
	if err != nil {
		log.Printf("cannot get port policy: %s %+v", responseBody, err)
		return models.PortPolicyResponse{}, err
	}

	var portPolicyResponse models.PortPolicyResponse
	err = json.Unmarshal(responseBody, &portPolicyResponse)
	if err != nil {
		log.Printf("cannot unmarshal port policy: %s %+v", responseBody, err)
		return models.PortPolicyResponse{}, err
	}

	return portPolicyResponse, nil
}

// deleteVLAN deletes VLAN
func deleteVLAN(vlan string) error {
	httpClient := httpclient.NewClient(models.BaseURL)

	responseBody, err := httpClient.SendRequest(context.Background(), "DELETE", fmt.Sprintf("/fabric/vlans/%s", vlan), nil)
	if err != nil {
		log.Printf("cannot delete VLAN: %s %+v", responseBody, err)
		return err
	}

	log.Printf("deleted VLAN %s", vlan)
	return nil
}

// DeletePortPolicy deletes port policy
func DeletePortPolicy(portPolicy string) error {
	httpClient := httpclient.NewClient(models.BaseURL)

	responseBody, err := httpClient.SendRequest(context.Background(), "DELETE", portPolicy, nil)
	if err != nil {
		log.Printf("cannot delete port policy: %s %+v", responseBody, err)
		return err
	}

	log.Printf("deleted port policy %s", portPolicy)
	return nil
}

// RemovePortPolicyFromEdgePort removes port policy from edge port
func RemovePortPolicyFromEdgePort(edgePort string, portPolicy string) error {

	port, err := fm.GetPort(edgePort)
	if err != nil {
		log.Printf("cannot get port details for port: %+v", err)
		return err
	}

	portPolicyLinks := port.PortPolicyLinks
	newPortPolicyLinks := []string{}
	for _, policy := range portPolicyLinks {
		if policy != portPolicy {
			newPortPolicyLinks = append(newPortPolicyLinks, policy)
		}
	}

	var portpolicylinksPATCHRequest models.PortPATCHRequest

	portpolicylinksPATCHRequest.PortPolicyLinks = newPortPolicyLinks

	httpClient := httpclient.NewClient(models.BaseURL)
	// send PATCH request to /fabric/ports/{edgePort} to remove port policy
	responseBody, err := httpClient.SendRequest(context.Background(), "PATCH", "/fabric/ports/"+edgePort, portpolicylinksPATCHRequest)
	if err != nil {
		log.Printf("cannot remove port policy from edge port: %s %+v", responseBody, err)
		return err
	}

	log.Printf("removed port policy %s from edge port %s", portPolicy, edgePort)
	return nil
}

// GetAllVNIPartitions gets all VNI partitions
func GetAllVNIPartitions() (models.AllVNIPartitionsResponse, error) {
	ctx := context.Background()

	var vniPartitions models.AllVNIPartitionsResponse
	httpClient := httpclient.NewClient(models.BaseURL)
	responseBody, err := httpClient.SendRequest(ctx, "GET", "/fabric/vni/partitions", nil)
	if err != nil {
		log.Printf("cannot get all VNI partitions: %s %+v", responseBody, err)
		return vniPartitions, err
	}

	err = json.Unmarshal(responseBody, &vniPartitions)
	if err != nil {
		log.Printf("cannot unmarshal all VNI partitions: %s %+v", responseBody, err)
		return vniPartitions, err
	}

	return vniPartitions, nil

}

// CheckVLANExists checks if VLAN exists
func CheckVLANExists(tenant *tapms.Tenant) (bool, int, error) {
	var vlanID int
	vlans, err := GetVLANs()
	if err != nil {
		log.Printf("cannot get VLANs: %+v", err)
		return false, vlanID, err
	}

	for _, vlan := range vlans {
		vlanID, err := strconv.Atoi(strings.Split(vlan, "/")[3])
		if err != nil {
			log.Printf("cannot convert VLAN ID to integer: %+v", err)
			return false, vlanID, err
		}

		vlan, err := GetVLAN(vlanID)
		if err != nil {
			log.Printf("cannot get VLAN: %+v", err)
			return false, vlanID, err
		}

		if vlan.VLANName == tenant.Spec.TenantName {
			log.Println("VLAN exists for tenant:", tenant.Spec.TenantName, vlan.VLANName)
			return true, vlanID, nil
		}
	}

	return false, vlanID, nil
}

// DeleteVLAN deletes VLAN
func DeleteVLAN(tenantName string, vlanID string) error {
	log.Println("deleting VLAN for tenant:", tenantName)

	//get all edge ports
	switches, err := fm.GetAllSwitches()
	if err != nil {
		return err
	}

	for _, x := range switches {
		DFAComponents, err := fm.GetSwitch(x)
		if err != nil {
			return err
		}

		for _, p := range DFAComponents.EdgePortsInfo {
			port, err := fm.GetPort(p.EdgePort)
			if err != nil {
				return err
			}

			for _, policy := range port.PortPolicyLinks {
				po := strings.Split(policy, "/")[3]
				if po == tenantName {
					err := RemovePortPolicyFromEdgePort(p.EdgePort, policy)
					if err != nil {
						log.Printf("cannot remove port policy from edge port: %+v", err)
					}
				}
			}

		}
	}

	err = DeletePortPolicy(fmt.Sprintf("/fabric/port-policies/%s", tenantName))
	if err != nil {
		log.Printf("cannot delete port policy: %+v", err)
		return err
	}

	//delete the VLAN
	err = deleteVLAN(vlanID)
	if err != nil {
		log.Printf("cannot delete VLAN: %+v", err)
		return err
	}

	return nil
}

// GetVlanID gets the VLAN ID for a VLAN
func GetVlanID(ctx context.Context, tenantName string) (int, error) {

	responseBody, err := httpClient.SendRequest(ctx, "GET", "/fabric/vlans", nil)
	if err != nil {
		log.Printf("cannot get VLANs: %s %+v", responseBody, err)
		return 0, err
	}

	var vlans models.VLANsResponse
	err = json.Unmarshal(responseBody, &vlans)
	if err != nil {
		log.Printf("cannot unmarshal VLANs: %s %+v", responseBody, err)
		return 0, err
	}

	for _, x := range vlans.DocumentLinks {
		vlanID, err := strconv.Atoi(strings.Split(x, "/")[3])
		if err != nil {
			log.Printf("cannot convert VLAN ID to integer: %+v", err)
			return 0, err
		}

		vlan, err := GetVLAN(vlanID)
		if err != nil {
			log.Printf("cannot get VLAN: %+v", err)
			return 0, err
		}

		if vlan.VLANName == tenantName {
			return vlanID, nil
		}
	}

	return 0, nil
}

// GetAllVNIBlocks gets all VNI blocks
func GetAllVNIBlocks() (models.AllVNIBlocksResponse, error) {
	ctx := context.Background()

	var vniBlocks models.AllVNIBlocksResponse
	responseBody, err := httpClient.SendRequest(ctx, "GET", "/fabric/vni/blocks", nil)
	if err != nil {
		log.Printf("cannot get all VNI blocks: %s %+v", responseBody, err)
		return vniBlocks, err
	}

	err = json.Unmarshal(responseBody, &vniBlocks)
	if err != nil {
		log.Printf("cannot unmarshal all VNI blocks: %s %+v", responseBody, err)
		return vniBlocks, err
	}

	return vniBlocks, nil

}

// CreateVNIBlock creates VNI block
func CreateVNIBlock(ctx context.Context, tenant tapms.Tenant, sshotTenant slingshot.SlingshotTenant) (models.VNIBlockResponse, error) {

	//Check if VNI block name is empty
	if sshotTenant.Spec.VNIBlockName == "" {
		return models.VNIBlockResponse{}, fmt.Errorf("VNI block name is empty")
	}

	//Get VNI Partition
	vniPartition, err := GetPartition(ctx, tenant.Spec.TenantName)
	if err != nil {
		log.Printf("cannot get VNI partition: %+v", err)
		return models.VNIBlockResponse{}, err
	}

	var vniBlockRequestData models.VNIBlockRequestData
	vniBlockRequestData.VNIPartitionName = tenant.Spec.TenantName
	vniBlockRequestData.VNIBlockName = fmt.Sprintf("%s-%s", tenant.Spec.TenantName, sshotTenant.Spec.VNIBlockName)
	vniBlockRequestData.VNIBlockRange = vniPartition.VNIRange

	var tenantXnames []string
	for _, t := range tenant.Spec.TenantResources {
		tenantXnames = append(tenantXnames, t.XNames...)
	}

	edgePortDFAList, _, err := GetEdgePortDFAList(tenantXnames)
	if err != nil {
		log.Printf("cannot get edge ports for tenant: %+v", err)
		return models.VNIBlockResponse{}, err
	}

	vniBlockRequestData.PortDFAs = edgePortDFAList

	// Send the request
	vniBlockResponseBody, err := httpClient.SendRequest(ctx, "POST", "/fabric/vni/blocks", vniBlockRequestData)
	if err != nil {
		log.Printf("cannot create VNI block: %s %+v", vniBlockResponseBody, err)
		return models.VNIBlockResponse{}, err
	}

	var vniBlockResponse models.VNIBlockResponse
	err = json.Unmarshal(vniBlockResponseBody, &vniBlockResponse)
	if err != nil {
		log.Printf("cannot unmarshal VNI block: %s %+v", vniBlockResponseBody, err)
		return models.VNIBlockResponse{}, err
	}

	return vniBlockResponse, nil
}

// CheckVniBlockEnforceTaskServiceState checks the state of VNI block enforcement task
func CheckVniBlockEnforceTaskServiceState(ctx context.Context, vniBlockEnforcementTaskServiceLink string) (bool, error) {
	log.Printf("checking the state of VNI block enforcement task %s", vniBlockEnforcementTaskServiceLink)
	var state models.VniBlockEnforcementTaskServiceState

	for {
		responseBody, err := httpClient.SendRequest(ctx, "GET", vniBlockEnforcementTaskServiceLink, nil)
		if err != nil {
			log.Printf("cannot get VNI block enforce task service state: %s %+v", responseBody, err)
			return false, err
		}

		err = json.Unmarshal(responseBody, &state)
		if err != nil {
			log.Printf("cannot unmarshal VNI block enforcement task service state: %s %+v", responseBody, err)
			return false, err
		}

		if state.TaskInfo.Stage == "FINISHED" || state.TaskInfo.Stage == "FAILED" {
			break
		}

		time.Sleep(models.WaitTime)
	}

	return state.TaskInfo.Stage == "FINISHED", nil

}

func ValidateVNIRequestData(vniRequestData models.VNIRequestData) error {
	if vniRequestData.VNICount < 0 || vniRequestData.VNICount > 65535 {
		return fmt.Errorf("VNI count is invalid: %d", vniRequestData.VNICount)
	}

	// the VNI Range is a slice of string
	// to compare it with integer value,
	// we need to split the string and convert it to integer
	// start should be greater than 0 and end should be less than 65536
	if len(vniRequestData.VNIRange) != 0 {
		vniRange := strings.Split(vniRequestData.VNIRange[0], "-")

		startRange, err := strconv.Atoi(vniRange[0])
		if err != nil {
			return fmt.Errorf("VNI range is invalid: %s", vniRange[0])
		}

		endRange, err := strconv.Atoi(vniRange[1])
		if err != nil {
			return fmt.Errorf("VNI range is invalid: %s", vniRange[1])
		}

		if startRange < 0 || endRange > 65536 || startRange > endRange {
			return fmt.Errorf("VNI range is invalid: %s", vniRequestData.VNIRange)
		}
	}

	return nil
}
